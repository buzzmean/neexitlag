; Build dependency only. Runtime uses Windows PowerShell 5.1, no SDK/Git/Go.
#ifndef StageRoot
 #define StageRoot "..\bin\installer"
#endif
#ifndef AppVersion
 #define AppVersion "0.3.0"
#endif
[Setup]
AppId={{E18D91DB-2E64-4405-A05E-AB15A4192318}
AppName=VELIQ
AppVersion={#AppVersion}
AppPublisher=VELIQ
DefaultDirName={autopf}\VELIQ
UsePreviousAppDir=no
SetupIconFile=..\brand\icons\app-icon.ico
UninstallDisplayIcon={app}\VELIQ.exe
AppPublisherURL=https://github.com/buzzmean/neexitlag
DisableDirPage=yes
DisableProgramGroupPage=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
OutputDir={#StageRoot}\output
OutputBaseFilename=VELIQ-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
; Brand images: dark side panel like app-ui/installer.html; regenerate with installer/make-wizard-images.py.
DisableWelcomePage=no
WizardImageFile=assets\wizard-100.bmp,assets\wizard-125.bmp,assets\wizard-150.bmp,assets\wizard-175.bmp,assets\wizard-200.bmp,assets\wizard-225.bmp,assets\wizard-250.bmp
WizardSmallImageFile=assets\wizard-small-100.bmp,assets\wizard-small-125.bmp,assets\wizard-small-150.bmp,assets\wizard-small-175.bmp,assets\wizard-small-200.bmp,assets\wizard-small-225.bmp,assets\wizard-small-250.bmp
WizardImageStretch=no
RestartIfNeededByRun=no
; Interactive installs close a running VELIQ.exe; the service-driven update passes
; /NOCLOSEAPPLICATIONS so a locked GUI is replaced at the next restart instead.
CloseApplications=yes
CloseApplicationsFilter=VELIQ.exe
RestartApplications=no
UninstallDisplayName=VELIQ
[Languages]
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"
[Messages]
WelcomeLabel1=Установка VELIQ
WelcomeLabel2=VELIQ строит оптимальный маршрут для каждого соединения: заблокированное открывается, игры получают самый быстрый путь.%n%nWindows один раз попросит права администратора — для службы, которая строит маршруты.
FinishedHeadingLabel=Готово
FinishedLabelNoIcons=VELIQ установлен. Откройте программу, введите код активации и нажмите «Оптимизировать».
FinishedLabel=VELIQ установлен. Откройте программу, введите код активации и нажмите «Оптимизировать».
[InstallDelete]
; Pre-rebrand installs used "VELIQ Setup"; its payload is superseded by {app}.
Type: filesandordirs; Name: "{autopf}\VELIQ Setup"
[Tasks]
Name: "desktopicon"; Description: "Создать ярлык на рабочем столе"; GroupDescription: "Ярлыки:"; Check: NotClubImage
[Files]
Source: "{#StageRoot}\payload\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#StageRoot}\payload\gui\VELIQ.exe"; DestDir: "{app}"; Flags: ignoreversion restartreplace
[Run]
Filename: "{sys}\WindowsPowerShell\v1.0\powershell.exe"; Parameters: "-NoProfile -NonInteractive -ExecutionPolicy Bypass -Command ""& '{app}\installer\install-package.ps1' -ClubCodeFile '{code:ClubCodeFile}' {code:ClubImageArgument} {code:UpdateArgument}"""; Flags: runhidden waituntilterminated; StatusMsg: "Устанавливаем VELIQ…"; BeforeInstall: ClearInstallMarker; AfterInstall: CheckInstall
Filename: "{app}\VELIQ.exe"; Description: "Открыть VELIQ и ввести код активации"; Flags: runasoriginaluser nowait postinstall skipifsilent; Check: NotClubImage
[Icons]
Name: "{autoprograms}\VELIQ"; Filename: "{app}\VELIQ.exe"
Name: "{autodesktop}\VELIQ"; Filename: "{app}\VELIQ.exe"; Tasks: desktopicon; Check: NotClubImage
[Code]
function ClubCodeFile(Param: String): String;
begin
 Result := ExpandConstant('{param:CLUBCODEFILE|}');
 if (Pos('''',Result)>0) or (Pos('"',Result)>0) or (Pos(#13,Result)>0) or (Pos(#10,Result)>0) then RaiseException('Недопустимый путь файла активации');
end;
function ClubImageArgument(Param: String): String;
begin
 Result := '';
 if ExpandConstant('{param:CLUBIMAGE|0}')='1' then Result := '-ClubImage';
end;
function UpdateArgument(Param: String): String;
begin
 Result := '';
 if ExpandConstant('{param:UPDATE|0}')='1' then Result := '-Update';
end;
function NotClubImage(): Boolean;
begin
 Result := ExpandConstant('{param:CLUBIMAGE|0}')<>'1';
end;
procedure ClearInstallMarker;
begin
 DeleteFile(ExpandConstant('{app}\install-success.json'));
end;
function InitializeUninstall(): Boolean;
var Code: Integer;
begin
 Result := Exec(ExpandConstant('{sys}\WindowsPowerShell\v1.0\powershell.exe'),
  '-NoProfile -NonInteractive -ExecutionPolicy Bypass -Command "& ' + #39 + ExpandConstant('{app}\installer\install-package.ps1') + #39 + ' -Remove"', '', SW_HIDE, ewWaitUntilTerminated, Code);
 Result := Result and (Code=0);
 if not Result then MsgBox('VELIQ: удаление остановлено. Данные сохранены; проверьте install-result.json.', mbError, MB_OK);
end;
procedure CheckInstall;
begin
 if not FileExists(ExpandConstant('{app}\install-success.json')) then
   RaiseException('Установка VELIQ не завершена. Состояние сохранено; код причины в install-result.json.');
end;
