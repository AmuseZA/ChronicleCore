; ChronicleCore Installer Script for Inno Setup
; Creates a single-EXE installer with everything embedded

#define MyAppName "ChronicleCore"
#define MyAppVersion "2.3.0"
#define MyAppPublisher "ChronicleCore"
#define MyAppExeName "ChronicleCore.bat"
#define MyAppURL "https://github.com/yourusername/chroniclecore"

[Setup]
; App identification
AppId={{8F7A3B2C-9D4E-4F1A-8B6C-7E5D3A2F1C0B}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir=installer_output
OutputBaseFilename=ChronicleCore_Setup_v{#MyAppVersion}
SetupIconFile=resources\app_icon.ico
UninstallDisplayIcon={app}\chroniclecore.exe
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
MinVersion=10.0.17763

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "startupicon"; Description: "Launch ChronicleCore on Windows startup"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
; Backend binary
Source: "dist_installer\chroniclecore.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "dist_installer\*.dll"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "dist_installer\web\*"; DestDir: "{app}\web"; Flags: ignoreversion recursesubdirs createallsubdirs

; Launcher script
Source: "dist_installer\ChronicleCore.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "dist_installer\debug_ml.bat"; DestDir: "{app}"; Flags: ignoreversion

; Embedded Python
Source: "dist_installer\python\*"; DestDir: "{app}\python"; Flags: ignoreversion recursesubdirs createallsubdirs

; ML sidecar
Source: "dist_installer\ml\*"; DestDir: "{app}\ml"; Flags: ignoreversion recursesubdirs createallsubdirs

; Database schema
Source: "dist_installer\spec\*"; DestDir: "{app}\spec"; Flags: ignoreversion recursesubdirs createallsubdirs

; Documentation
Source: "dist_installer\QUICK_START.md"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "dist_installer\README_USER.md"; DestDir: "{app}"; Flags: ignoreversion isreadme skipifsourcedoesntexist

[Icons]
; Desktop shortcut
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon; IconFilename: "{app}\chroniclecore.exe"; Comment: "Launch ChronicleCore"

; Start menu shortcut
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\chroniclecore.exe"; Comment: "Launch ChronicleCore"

; Startup shortcut (optional)
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: startupicon; IconFilename: "{app}\chroniclecore.exe"; Comment: "Launch ChronicleCore"

; Uninstaller
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

[Run]
; Offer to launch after install
Filename: "{app}\{#MyAppExeName}"; Description: "Launch ChronicleCore"; Flags: postinstall nowait skipifsilent shellexec

[Code]
var
  ProgressPage: TOutputProgressWizardPage;

procedure InitializeWizard;
begin
  ProgressPage := CreateOutputProgressPage('Installing', 'Please wait while Setup installs ChronicleCore on your computer.');
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    ProgressPage.SetText('Verifying installation...', '');
    ProgressPage.SetProgress(90, 100);

    // Verify Python exists
    if not FileExists(ExpandConstant('{app}\python\python.exe')) then
    begin
      MsgBox('Warning: Python runtime not found. Installation may be incomplete.', mbError, MB_OK);
    end;

    ProgressPage.SetProgress(100, 100);
  end;
end;

function InitializeUninstall(): Boolean;
var
  Response: Integer;
begin
  Response := MsgBox('Do you want to keep your data (tracked time, profiles, etc.)?'#13#10#13#10 +
                     'Your data is stored in:'#13#10 +
                     ExpandConstant('{localappdata}\ChronicleCore\chronicle.db'),
                     mbConfirmation, MB_YESNO);

  if Response = IDYES then
  begin
    MsgBox('Your data will be preserved. You can find it in:'#13#10 +
           ExpandConstant('{localappdata}\ChronicleCore\'),
           mbInformation, MB_OK);
  end
  else
  begin
    // Delete data directory
    DelTree(ExpandConstant('{localappdata}\ChronicleCore'), True, True, True);
  end;

  Result := True;
end;

[UninstallDelete]
Type: filesandordirs; Name: "{app}"

[Registry]
; Store installation info for updates
Root: HKCU; Subkey: "Software\ChronicleCore"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\ChronicleCore"; ValueType: string; ValueName: "Version"; ValueData: "{#MyAppVersion}"; Flags: uninsdeletekey











































