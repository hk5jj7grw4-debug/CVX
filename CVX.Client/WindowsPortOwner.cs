using System.ComponentModel;
using System.Runtime.InteropServices;

namespace CVX.Client;

internal static class WindowsPortOwner
{
    // The kernel is probed over IPv4 loopback. TCP_TABLE_OWNER_PID_LISTENER = 3.
    [DllImport("iphlpapi.dll")]
    static extern uint GetExtendedTcpTable(IntPtr table, ref int size, bool order, int family, int tableClass, uint reserved);

    internal static int? Find(int port)
    {
        if (!OperatingSystem.IsWindows()) return null;
        var size = 0;
        var code = GetExtendedTcpTable(IntPtr.Zero, ref size, false, 2, 3, 0);
        if (code != 122 && code != 0) throw new Win32Exception((int)code);
        for (var attempt = 0; attempt < 4; attempt++)
        {
            var memory = Marshal.AllocHGlobal(size);
            try
            {
                code = GetExtendedTcpTable(memory, ref size, false, 2, 3, 0);
                if (code == 122) continue;
                if (code != 0) throw new Win32Exception((int)code);
                var count = Marshal.ReadInt32(memory);
                for (var i = 0; i < count; i++)
                {
                    var row = IntPtr.Add(memory, 4 + i * 24);
                    var address = unchecked((uint)Marshal.ReadInt32(row, 4));
                    var localPort = (Marshal.ReadByte(row, 8) << 8) | Marshal.ReadByte(row, 9);
                    if (localPort == port && (address == 0 || address == 0x0100007f))
                        return Marshal.ReadInt32(row, 20);
                }
                return null;
            }
            finally { Marshal.FreeHGlobal(memory); }
        }
        throw new InvalidOperationException("无法稳定读取端口所属进程，请重试");
    }
}
