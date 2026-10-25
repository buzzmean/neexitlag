using System.Windows;
using System.Windows.Controls;
namespace Neexitlag.Desktop;
public partial class GamesView:UserControl {
    public GamesView()=>InitializeComponent();
    void SelectGame(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel vm && sender is Button {DataContext:GameCard game})vm.Select(game);}
    void ChooseRoute(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel vm && sender is Button {DataContext:RouteOption route})vm.ChooseRoute(route);}
    async void Apply(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel {Selected:{} game} vm)await vm.ApplyAsync(game);}
    async void Repair(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel {Selected:{} game} vm)await vm.RepairLauncherAsync(game);}
}
