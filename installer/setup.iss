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
DefaultDirName={autopf}\VELIQ Setup
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
RestartIfNeededByRun=no
CloseApplications=no
RestartApplications=no
UninstallDisplayName=VELIQ
[Languages]
Name: "russian"; MessagesFile: "compiler:Languages\Russian.isl"
[Files]
Source: "{#StageRoot}\payload\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
[Run]
Filename: "{sys}\WindowsPowerShell\v1.0\powershell.exe"; Parameters: "-NoProfile -NonInteractive -ExecutionPolicy Bypass -Command ""& '{app}\installer\install-package.ps1' -ClubCodeFile '{code:ClubCodeFile}'"""; Flags: runhidden waituntilterminated; StatusMsg: "Устанавливаем VELIQ…"; BeforeInstall: ClearInstallMarker; AfterInstall: CheckInstall
Filename: "{localappdata}\Programs\VELIQ\VELIQ.exe"; Description: "Открыть VELIQ и ввести код активации"; Flags: runasoriginaluser nowait skipifsilent
[Code]
function ClubCodeFile(Param: String): String;
begin
 Result := ExpandConstant('{param:CLUBCODEFILE|}');
 if (Pos('''',Result)>0) or (Pos('"',Result)>0) or (Pos(#13,Result)>0) or (Pos(#10,Result)>0) then RaiseException('Недопустимый путь файла активации');
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
