using System.Windows;
using System.Windows.Controls;
namespace Neexitlag.Desktop;
public partial class GamesView:UserControl {
    public GamesView()=>InitializeComponent();
    async void Apply(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel vm && sender is Button {DataContext:GameCard game})await vm.ApplyAsync(game);}
    async void Repair(object sender,RoutedEventArgs e){if(DataContext is GameCatalogViewModel vm && sender is Button {DataContext:GameCard game})await vm.RepairLauncherAsync(game);}
}
