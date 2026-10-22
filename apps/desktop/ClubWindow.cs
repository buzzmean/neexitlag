using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Threading;
using System.Text.Json;
namespace Neexitlag.Desktop;
// Club UI deliberately has no settings, activation, disable or routing mutations.
// Actual authorization is enforced by the service, independently of this window.
public sealed class ClubWindow : Window
{
    readonly System.Windows.Forms.NotifyIcon tray=new();
    readonly DispatcherTimer timer=new(){Interval=TimeSpan.FromSeconds(5)};
    readonly TextBlock status=new(){Text="Подготавливаем VELIQ…",FontSize=20,TextWrapping=TextWrapping.Wrap,Margin=new Thickness(24)};
    readonly IServiceClient service=new ServiceClient();bool polling,exiting;
    public ClubWindow()
    {
        Title="VELIQ";Width=440;Height=190;Background=Brushes.Black;Foreground=Brushes.White;Content=status;
        var stream=System.Windows.Application.GetResourceStream(new Uri("pack://application:,,,/Assets/NeExitLag.ico"));
        if(stream is not null)tray.Icon=new System.Drawing.Icon(stream.Stream);
        tray.Text="VELIQ · клуб";tray.Visible=true;
        var menu=new System.Windows.Forms.ContextMenuStrip();menu.Items.Add("Состояние VELIQ",null,(_,_)=>Dispatcher.Invoke(ShowFromTray));tray.ContextMenuStrip=menu;
        tray.DoubleClick+=(_,_)=>Dispatcher.Invoke(ShowFromTray);
        Closing+=(_,e)=>{if(!exiting){e.Cancel=true;Hide();}};
        timer.Tick+=async(_,_)=>await Poll();timer.Start();_=Poll();
    }
    async Task Poll()
    {
        if(polling)return;polling=true;
        try {var s=await service.Call<JsonElement>("club.status");var state=s.GetProperty("state").GetString();status.Text=state=="BASELINE_PASS_RUNTIME_UNVERIFIED" ? "VELIQ · подключение готово\nМаршрутизация новых соединений включена" : state=="ACTIVATION_ERROR" ? "VELIQ · требуется внимание администратора" : "VELIQ · подготовка подключения…";}
        catch {status.Text="VELIQ · ожидаем готовность службы…";}
        finally {polling=false;}
    }
    public void ShowFromTray(){Show();WindowState=WindowState.Normal;Activate();}
    // Same-owner installer activation pipe only; never disables service or browsers.
    public void Exit(){exiting=true;timer.Stop();tray.Visible=false;tray.Dispose();Close();System.Windows.Application.Current.Shutdown();}
    public static bool IsClubInstallation()
    {
        try{return ClubPresentation.Enabled(File.ReadAllText(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.CommonApplicationData),"neexitlag","club-mode.json")));}
        catch(IOException){return false;}catch(UnauthorizedAccessException){return false;}
    }
}
