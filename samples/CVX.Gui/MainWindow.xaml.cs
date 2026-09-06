using System.Collections.ObjectModel;
using System.ComponentModel;
using System.IO;
using System.Net.Http;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using CVX.Client;
using CVX.Sdk;
using Microsoft.Win32;
using CVX.Gui.Services;

namespace CVX.Gui;

public partial class MainWindow : Window
{
    private const int MaxLogItems = 400;
    private const string CallbackUrl = "http://127.0.0.1:5000/api/recvMsg";

    private readonly ObservableCollection<ChatLogItem> _logItems = [];
    private readonly ObservableCollection<RoomOption> _rooms = [];
    private readonly SolidColorBrush _okBrush = BrushFrom("#07C160");
    private readonly SolidColorBrush _busyBrush = BrushFrom("#F5A524");
    private readonly SolidColorBrush _errorBrush = BrushFrom("#E64545");
    private readonly HttpClient _http = new();
    private readonly Dictionary<string, ImageSource> _avatarCache = new(StringComparer.OrdinalIgnoreCase);

    private WechatRobotRuntime? _runtime;
    private CVXClient? _sdk;

    private ImageSource? _selfAvatar;
    private string? _selfAvatarUrl;
    private bool _busy;
    private bool _closing;
    private bool _kernelReady;
    private WechatRuntimePhase _lastPhase;
    private bool _lastLoggedIn;

    public MainWindow()
    {
        InitializeComponent();
        MessageList.ItemsSource = _logItems;
        RoomBox.ItemsSource = _rooms;
        TargetBox.Items.Add("filehelper");
        _http.DefaultRequestHeaders.UserAgent.ParseAdd("Mozilla/5.0");
        ApplyCredentialsToForm(CredentialStore.Load());
        GvxApiText.Text = $"127.0.0.1:{ReadCredentialsFromForm().GvxApiPort}";
        HeaderStatusText.Text = "请填写配置后接入";

        Loaded += async (_, _) => await InitializeClientAsync();
        Closing += OnWindowClosing;
    }

    private async Task InitializeClientAsync()
    {
        var credentials = ReadCredentialsFromForm();
        if (string.IsNullOrWhiteSpace(credentials.UpdateServer)
            && string.IsNullOrWhiteSpace(credentials.ComponentToken))
        {
            HeaderStatusText.Text = "请先填写组件服务 URL 和 Token";
            return;
        }

        await ConnectKernelAsync();
    }

    private void SaveConfigButton_Click(object sender, RoutedEventArgs e)
    {
        var credentials = ReadCredentialsFromForm();
        CredentialStore.Save(credentials);
        GvxApiText.Text = $"127.0.0.1:{credentials.GvxApiPort}";
        AppendLog("sys", "系统", "配置已保存。");
        HeaderStatusText.Text = "配置已保存，可点重新接入";
    }

    private AppCredentials ReadCredentialsFromForm()
    {
        var port = 19088;
        if (int.TryParse(GvxPortBox.Text?.Trim(), out var parsed) && parsed is > 0 and <= 65535)
        {
            port = parsed;
        }

        return new AppCredentials
        {
            UpdateServer = UpdateServerBox.Text?.Trim() ?? "",
            ComponentToken = ComponentTokenBox.Password?.Trim() ?? "",
            GvxApiPort = port
        };
    }

    private void ApplyCredentialsToForm(AppCredentials credentials)
    {
        UpdateServerBox.Text = credentials.UpdateServer;
        ComponentTokenBox.Password = credentials.ComponentToken;
        GvxPortBox.Text = credentials.GvxApiPort <= 0 ? "19088" : credentials.GvxApiPort.ToString();
    }

    private void EnsureRuntime()
    {
        if (_runtime is not null)
        {
            return;
        }

        var credentials = ReadCredentialsFromForm();
        CredentialStore.Save(credentials);
        GvxApiText.Text = $"127.0.0.1:{credentials.GvxApiPort}";
        _sdk = new CVXClient($"http://127.0.0.1:{credentials.GvxApiPort}");
        _runtime = new WechatRobotRuntime(new WechatRobotOptions
        {
            UpdateServerUrl = string.IsNullOrWhiteSpace(credentials.UpdateServer) ? null : credentials.UpdateServer,
            ComponentToken = string.IsNullOrWhiteSpace(credentials.ComponentToken) ? null : credentials.ComponentToken,
            GvxApiPort = credentials.GvxApiPort,
            CallbackUrl = CallbackUrl
        });

        _runtime.MessageReceived += (_, message) =>
        {
            Dispatcher.BeginInvoke(() =>
            {
                var senderName = string.IsNullOrWhiteSpace(message.SenderName)
                    ? message.SenderId
                    : message.SenderName;
                var prefix = message.IsGroupMessage ? "[群] " : "";
                AppendLog("recv", $"{prefix}{senderName}", message.Text, message.SenderId);
            });
        };

        _runtime.ReceiveError += (_, error) =>
        {
            Dispatcher.BeginInvoke(() =>
            {
                LastErrorText.Text = error.Message;
                AppendLog("error", "回调错误", error.Message);
            });
        };

        _runtime.StatusChanged += OnRuntimeStatusChanged;
        RenderStatus(_runtime.Status);
    }

    private void OnRuntimeStatusChanged(object? sender, WechatRobotStatus status)
    {
        if (_closing)
        {
            return;
        }

        Dispatcher.BeginInvoke(new Action(() =>
        {
            if (!_closing)
            {
                RenderStatus(status);
            }
        }));
    }

    private async void StartButton_Click(object sender, RoutedEventArgs e)
    {
        await ConnectKernelAsync();
    }

    private async void RestartButton_Click(object sender, RoutedEventArgs e)
    {
        await RestartKernelAsync();
    }

    private async Task ConnectKernelAsync()
    {
        if (_busy)
        {
            return;
        }

        try
        {
            SetBusy(true);
            EnsureRuntime();
            await _runtime!.ConnectAsync();
        }
        catch (Exception)
        {
            if (_runtime is not null)
            {
                RenderStatus(_runtime.Status);
            }
        }
        finally
        {
            SetBusy(false);
        }
    }

    private async Task RestartKernelAsync()
    {
        if (_busy)
        {
            return;
        }

        try
        {
            SetBusy(true);
            EnsureRuntime();
            ClearAccount();
            await _runtime!.RestartAsync();
        }
        catch (Exception)
        {
            if (_runtime is not null)
            {
                RenderStatus(_runtime.Status);
            }
        }
        finally
        {
            SetBusy(false);
        }
    }

    private async void RefreshRoomsButton_Click(object sender, RoutedEventArgs e)
    {
        if (!_kernelReady)
        {
            AppendLog("error", "错误", "请先连接内核。");
            return;
        }

        try
        {
            SetBusy(true);
            var rooms = await _sdk!.Room.GetChatroomListAsync();
            _rooms.Clear();
            foreach (var room in rooms)
            {
                var option = new RoomOption(room.GroupId, room.GroupName, room.Remark);
                _rooms.Add(option);
                if (!TargetBox.Items.Contains(room.GroupId) && !string.IsNullOrWhiteSpace(room.GroupId))
                {
                    TargetBox.Items.Add(room.GroupId);
                }
            }

            AppendLog("sys", "系统", $"已加载 {_rooms.Count} 个群聊。");
        }
        catch (Exception ex)
        {
            AppendLog("error", "错误", ex.Message);
        }
        finally
        {
            SetBusy(false);
        }
    }

    private void ClearLogButton_Click(object sender, RoutedEventArgs e)
    {
        _logItems.Clear();
    }

    private void RoomBox_SelectionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (RoomBox.SelectedItem is RoomOption room && !string.IsNullOrWhiteSpace(room.Id))
        {
            TargetBox.Text = room.Id;
        }
    }

    private async void SendButton_Click(object sender, RoutedEventArgs e)
    {
        if (!_kernelReady)
        {
            AppendLog("error", "错误", "请先连接内核。");
            return;
        }

        var target = TargetBox.Text?.Trim();
        if (string.IsNullOrWhiteSpace(target))
        {
            AppendLog("error", "错误", "请填写发送对象。");
            return;
        }

        var sendType = (SendTypeBox.SelectedItem as ComboBoxItem)?.Content?.ToString() ?? "文本";
        var content = SendContentBox.Text?.Trim() ?? "";

        try
        {
            SetBusy(true);
            switch (sendType)
            {
                case "图片":
                    var imagePath = PickFile("图片|*.png;*.jpg;*.jpeg;*.gif;*.bmp|所有文件|*.*");
                    if (imagePath is null)
                    {
                        return;
                    }

                    await _sdk!.Message.SendImageAsync(target, imagePath);
                    AppendLog("send", "我", $"[图片] {imagePath} → {target}", target);
                    break;
                case "文件":
                    var filePath = PickFile("所有文件|*.*");
                    if (filePath is null)
                    {
                        return;
                    }

                    await _sdk!.Message.SendFileAsync(target, filePath);
                    AppendLog("send", "我", $"[文件] {filePath} → {target}", target);
                    break;
                case "拍一拍":
                    var patWxid = string.IsNullOrWhiteSpace(content) ? target : content;
                    await _sdk!.Message.SendPatAsync(target, patWxid);
                    AppendLog("send", "我", $"[拍一拍] {patWxid} @ {target}", target);
                    break;
                default:
                    if (string.IsNullOrWhiteSpace(content))
                    {
                        AppendLog("error", "错误", "请输入要发送的文本。");
                        return;
                    }

                    await _sdk!.Message.SendTextAsync(target, content);
                    AppendLog("send", "我", content, target);
                    SendContentBox.Clear();
                    break;
            }
        }
        catch (Exception ex)
        {
            AppendLog("error", "错误", ex.Message);
        }
        finally
        {
            SetBusy(false);
        }
    }

    private void RenderStatus(WechatRobotStatus status)
    {
        _kernelReady = status.ApiReady && status.Phase == WechatRuntimePhase.Connected;
        PhaseText.Text = PhaseLabel(status.Phase);
        ApiReadyText.Text = status.ApiReady ? "已就绪" : "未就绪";
        LoginStateText.Text = status.ApiReady
            ? status.IsLoggedIn ? "已登录" : "未登录"
            : "—";
        GvxApiText.Text = $"127.0.0.1:{ReadCredentialsFromForm().GvxApiPort}";
        CallbackReadyText.Text = status.CallbackUrl.ToString();
        ComponentVersionText.Text = string.IsNullOrWhiteSpace(status.ComponentVersion)
            ? "—"
            : status.ComponentVersion;
        RequiredWechatText.Text = string.IsNullOrWhiteSpace(status.RequiredWechatVersion)
            ? "—"
            : status.RequiredWechatVersion;
        DetectedWechatText.Text = string.IsNullOrWhiteSpace(status.DetectedWechatVersion)
            ? "未知"
            : status.DetectedWechatVersion;
        VersionMatchText.Text = status.VersionMatch switch
        {
            WechatVersionMatch.Matching => "匹配",
            WechatVersionMatch.Mismatched => "不匹配",
            _ => "未知"
        };
        CheckedAtText.Text = status.CheckedAt is { } checkedAt
            ? checkedAt.ToLocalTime().ToString("HH:mm:ss")
            : "—";
        LastErrorText.Text = string.IsNullOrWhiteSpace(status.Error) ? "无" : status.Error;

        HeaderStatusText.Text = HeaderText(status);
        StatusDot.Background = status.Phase switch
        {
            WechatRuntimePhase.Connected when status.IsLoggedIn => _okBrush,
            WechatRuntimePhase.Failed or WechatRuntimePhase.Disposed => _errorBrush,
            WechatRuntimePhase.Connected => _busyBrush,
            _ => _busyBrush
        };

        if (status.DownloadProgress is { } progress)
        {
            DownloadProgress.IsIndeterminate = false;
            DownloadProgress.Value = Math.Clamp(progress, 0, 100);
            DownloadProgressText.Text = $"下载进度：{(int)Math.Round(progress)}%";
        }
        else if (status.Phase is WechatRuntimePhase.Downloading or WechatRuntimePhase.Repairing
                 or WechatRuntimePhase.Checking or WechatRuntimePhase.Connecting)
        {
            DownloadProgress.IsIndeterminate = true;
            DownloadProgressText.Text = PhaseLabel(status.Phase);
        }
        else
        {
            DownloadProgress.IsIndeterminate = false;
            DownloadProgress.Value = status.Phase == WechatRuntimePhase.Connected ? 100 : 0;
            DownloadProgressText.Text = status.Phase == WechatRuntimePhase.Connected ? "内核已接入" : "等待连接";
        }

        if (status.Phase != _lastPhase)
        {
            if (status.Phase == WechatRuntimePhase.Connected)
            {
                NativeRuntimeFixer.EnsureVcRuntimeBesideWechat();
                AppendLog(
                    "sys",
                    "系统",
                    status.IsLoggedIn
                        ? $"已连接，微信已登录。回调={status.CallbackUrl}"
                        : $"内核已就绪，请登录微信。回调={status.CallbackUrl}");
            }
            else if (status.Phase == WechatRuntimePhase.Failed && !string.IsNullOrWhiteSpace(status.Error))
            {
                AppendLog("error", "错误", status.Error);
            }

            _lastPhase = status.Phase;
        }

        if (status.Phase == WechatRuntimePhase.Connected && status.IsLoggedIn && !_lastLoggedIn)
        {
            _ = RefreshAccountAsync();
        }
        else if (status.Phase == WechatRuntimePhase.Connected && !status.IsLoggedIn)
        {
            AccountText.Text = "等待微信登录";
        }

        _lastLoggedIn = status.Phase == WechatRuntimePhase.Connected && status.IsLoggedIn;
    }

    private static string PhaseLabel(WechatRuntimePhase phase) => phase switch
    {
        WechatRuntimePhase.Idle => "尚未连接",
        WechatRuntimePhase.Checking => "正在检查微信版本",
        WechatRuntimePhase.Downloading => "正在下载组件",
        WechatRuntimePhase.Repairing => "正在恢复微信",
        WechatRuntimePhase.Connecting => "正在启动并连接内核",
        WechatRuntimePhase.Connected => "已连接",
        WechatRuntimePhase.Failed => "失败",
        WechatRuntimePhase.Disposed => "已停止监控",
        _ => phase.ToString()
    };

    private static string HeaderText(WechatRobotStatus status) => status.Phase switch
    {
        WechatRuntimePhase.Connected when status.IsLoggedIn => "已连接，微信已登录",
        WechatRuntimePhase.Connected => "内核已就绪，请登录微信",
        WechatRuntimePhase.Failed => string.IsNullOrWhiteSpace(status.Error)
            ? "连接异常，可点重新接入"
            : status.Error,
        WechatRuntimePhase.Disposed => "已停止监控",
        _ => PhaseLabel(status.Phase)
    };

    private async Task RefreshAccountAsync()
    {
        if (_sdk is null)
        {
            return;
        }

        try
        {
            var login = await _sdk.System.GetLoginStatusAsync();
            if (!login.Data.Status)
            {
                AccountText.Text = "等待微信登录";
                return;
            }

            string? nick = null;
            var wxid = login.AccountWxid;
            string? avatarUrl = null;
            try
            {
                var profile = await _sdk!.Contact.GetProfileNewAsync();
                nick = profile.UserInfo?.NickName?.Value;
                wxid = profile.UserInfo?.UserName?.Value ?? wxid;
                avatarUrl = profile.UserInfoExt?.SmallHeadImgUrl
                            ?? profile.UserInfoExt?.BigHeadImgUrl;
            }
            catch (CVXApiException)
            {
                // 刚登录时资料接口可能还没就绪，wxid 已经足够显示。
            }

            AccountText.Text = string.IsNullOrWhiteSpace(nick)
                ? wxid
                : string.IsNullOrWhiteSpace(wxid) ? nick : $"{nick}  ({wxid})";

            if (!string.IsNullOrWhiteSpace(wxid))
            {
                await ApplySelfAvatarAsync(wxid, avatarUrl);
            }

            if (HasAccountInfo())
            {
                AppendLog("sys", "系统", $"当前微信：{AccountText.Text}");
            }
        }
        catch (Exception)
        {
            AccountText.Text = "等待微信登录";
        }
    }

    private bool HasAccountInfo()
    {
        var text = AccountText.Text;
        return !string.IsNullOrWhiteSpace(text)
               && text != "等待微信登录";
    }

    private void ClearAccount()
    {
        AccountText.Text = "";
        AccountAvatar.Visibility = Visibility.Collapsed;
        _selfAvatar = null;
        _selfAvatarUrl = null;
        _lastLoggedIn = false;
    }

    private void SetBusy(bool busy)
    {
        _busy = busy;
        StartButton.IsEnabled = !busy;
        RestartButton.IsEnabled = !busy;
        SendButton.IsEnabled = !busy;
        RefreshRoomsButton.IsEnabled = !busy;
        SaveConfigButton.IsEnabled = !busy;
        if (busy)
        {
            StatusDot.Background = _busyBrush;
        }
    }

    private void AppendLog(string kind, string sender, string? text, string? wxid = null)
    {
        var item = new ChatLogItem
        {
            Kind = kind,
            Sender = sender,
            Time = DateTime.Now.ToString("HH:mm:ss"),
            Text = string.IsNullOrWhiteSpace(text) ? "（空消息）" : text,
            Avatar = kind == "send" ? _selfAvatar : null
        };
        _logItems.Add(item);

        while (_logItems.Count > MaxLogItems)
        {
            _logItems.RemoveAt(0);
        }

        if (MessageList.Items.Count > 0)
        {
            MessageList.ScrollIntoView(MessageList.Items[^1]);
        }

        if (kind == "recv" && !string.IsNullOrWhiteSpace(wxid))
        {
            _ = BindAvatarAsync(item, wxid);
        }
    }

    private async Task ApplySelfAvatarAsync(string wxid, string? avatarUrl)
    {
        if (_selfAvatar is not null && string.Equals(_selfAvatarUrl, avatarUrl, StringComparison.Ordinal))
        {
            AccountAvatar.Visibility = Visibility.Visible;
            return;
        }

        var avatar = await ResolveAvatarAsync(wxid, avatarUrl);
        if (avatar is null)
        {
            return;
        }

        _selfAvatar = avatar;
        _selfAvatarUrl = avatarUrl;
        AccountAvatarBrush.ImageSource = avatar;
        AccountAvatar.Visibility = Visibility.Visible;
    }

    private async Task BindAvatarAsync(ChatLogItem item, string wxid)
    {
        try
        {
            var avatar = await ResolveAvatarAsync(wxid, null);
            if (avatar is not null)
            {
                item.Avatar = avatar;
            }
        }
        catch (Exception)
        {
            // Ignore avatar lookup failures; the message text still shows.
        }
    }

    private async Task<ImageSource?> ResolveAvatarAsync(string wxid, string? avatarUrl)
    {
        if (_avatarCache.TryGetValue(wxid, out var cached))
        {
            return cached;
        }

        var url = avatarUrl;
        if (string.IsNullOrWhiteSpace(url) && _sdk is not null)
        {
            try
            {
                var contact = await _sdk.Contact.GetContactFastAsync(wxid);
                url = contact.Contact?.SmallHeadImgUrl ?? contact.Contact?.BigHeadImgUrl;
                var fromBuffer = DecodeImageBuffer(contact.Contact?.ImgBuf?.Buffer);
                if (fromBuffer is not null)
                {
                    _avatarCache[wxid] = fromBuffer;
                    return fromBuffer;
                }
            }
            catch (Exception)
            {
                // Fall back to the URL if the contact API is not ready.
            }
        }

        var downloaded = await DownloadAvatarAsync(url);
        if (downloaded is not null)
        {
            _avatarCache[wxid] = downloaded;
        }

        return downloaded;
    }

    private async Task<ImageSource?> DownloadAvatarAsync(string? url)
    {
        if (string.IsNullOrWhiteSpace(url))
        {
            return null;
        }

        try
        {
            using var response = await _http.GetAsync(url);
            response.EnsureSuccessStatusCode();
            var bytes = await response.Content.ReadAsByteArrayAsync();
            return CreateImage(bytes);
        }
        catch (Exception)
        {
            return null;
        }
    }

    private static ImageSource? DecodeImageBuffer(string? buffer)
    {
        if (string.IsNullOrWhiteSpace(buffer))
        {
            return null;
        }

        try
        {
            return CreateImage(Convert.FromBase64String(buffer));
        }
        catch (FormatException)
        {
            return null;
        }
    }

    private static ImageSource? CreateImage(byte[] bytes)
    {
        if (bytes.Length == 0)
        {
            return null;
        }

        var image = new BitmapImage();
        using var stream = new MemoryStream(bytes);
        image.BeginInit();
        image.CacheOption = BitmapCacheOption.OnLoad;
        image.StreamSource = stream;
        image.EndInit();
        image.Freeze();
        return image;
    }

    private static string? PickFile(string filter)
    {
        var dialog = new OpenFileDialog
        {
            Filter = filter,
            CheckFileExists = true
        };
        return dialog.ShowDialog() == true ? dialog.FileName : null;
    }

    private async void OnWindowClosing(object? sender, CancelEventArgs e)
    {
        if (_closing)
        {
            return;
        }

        e.Cancel = true;
        _closing = true;
        if (_runtime is not null)
        {
            _runtime.StatusChanged -= OnRuntimeStatusChanged;
            try
            {
                await _runtime.DisposeAsync();
            }
            catch (Exception)
            {
                // Closing still proceeds so the window cannot get stuck.
            }
        }

        _sdk?.Dispose();
        _http.Dispose();
        Close();
    }

    private static SolidColorBrush BrushFrom(string hex)
    {
        var brush = new SolidColorBrush((Color)ColorConverter.ConvertFromString(hex));
        brush.Freeze();
        return brush;
    }

    public sealed class ChatLogItem : INotifyPropertyChanged
    {
        private ImageSource? _avatar;

        public string Kind { get; init; } = "recv";
        public string Sender { get; init; } = "";
        public string Time { get; init; } = "";
        public string Text { get; init; } = "";
        public bool ShowAvatar => Kind is "recv" or "send";

        public ImageSource? Avatar
        {
            get => _avatar;
            set
            {
                _avatar = value;
                PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(Avatar)));
            }
        }

        public event PropertyChangedEventHandler? PropertyChanged;
    }

    public sealed class RoomOption(string? id, string? name, string? remark)
    {
        public string Id { get; } = id ?? "";
        public string Display { get; } =
            string.IsNullOrWhiteSpace(name) ? id ?? "" :
            string.IsNullOrWhiteSpace(remark) ? $"{name}  ({id})" : $"{name} / {remark}  ({id})";
    }
}
