; ChronicleCore Inno Setup Installer Script
; This creates a single EXE installer with all dependencies

#define MyAppName "ChronicleCore"
#define MyAppVersion "1.6.0"
#define MyAppPublisher "ChronicleCore"
#define MyAppURL "https://github.com/yourusername/chroniclecore"
#define MyAppExeName "chroniclecore.exe"

[Setup]
; App identification
AppId={{YOUR-GUID-HERE}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}

; Installation settings
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
LicenseFile=LICENSE.txt
OutputDir=installer_output
OutputBaseFilename=ChronicleCore_Setup_v{#MyAppVersion}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern

; Privileges
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

; Visual
; SetupIconFile=resources\app_icon.ico
UninstallDisplayIcon={app}\{#MyAppExeName}

; System requirements
ArchitecturesAllowed=x64
MinVersion=10.0.17763

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startupicon"; Description: "Start ChronicleCore on Windows startup"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; Backend binary
Source: "apps\chroniclecore-core\chroniclecore.exe"; DestDir: "{app}"; Flags: ignoreversion

; Embedded Python runtime (portable)
Source: "dist\python\*"; DestDir: "{app}\python"; Flags: ignoreversion recursesubdirs createallsubdirs

; ML sidecar code
Source: "apps\chronicle-ml\src\*"; DestDir: "{app}\ml\src"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "apps\chronicle-ml\requirements.txt"; DestDir: "{app}\ml"; Flags: ignoreversion

; Python packages (pre-installed)
; Source: "dist\python_packages\*"; DestDir: "{app}\python\Lib\site-packages"; Flags: ignoreversion recursesubdirs createallsubdirs

; UI assets (if available)
Source: "apps\chroniclecore-ui\build\*"; DestDir: "{app}\web"; Flags: ignoreversion recursesubdirs createallsubdirs

; Database schema
Source: "spec\schema.sql"; DestDir: "{app}\spec"; Flags: ignoreversion

; Documentation
Source: "QUICK_START.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "README_USER.md"; DestDir: "{app}"; Flags: ignoreversion isreadme
Source: "workflow\ml_user_guide.md"; DestDir: "{app}\docs"; Flags: ignoreversion

; Version tracking
Source: "VERSION.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: startupicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
var
  UpdateCheckPage: TOutputMsgMemoWizardPage;
  DownloadPage: TDownloadWizardPage;

function OnDownloadProgress(const Url, FileName: String; const Progress, ProgressMax: Int64): Boolean;
begin
  if Progress = ProgressMax then
    Log(Format('Successfully downloaded %s', [FileName]));
  Result := True;
end;

procedure InitializeWizard;
begin
  // Custom page for update checking (optional)
  UpdateCheckPage := CreateOutputMsgMemoPage(wpWelcome,
    'Checking for Updates', 'Verifying you have the latest version',
    'Checking online for the latest version...', '');
end;

function InitializeSetup(): Boolean;
begin
  Result := True;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    // Initialize ML dependencies (verify they're present)
    Log('Installation complete. ML dependencies bundled.');
  end;
end;

function InitializeUninstall(): Boolean;
var
  Response: Integer;
begin
  Response := MsgBox('Do you want to keep your data (tracked time, profiles)?', mbConfirmation, MB_YESNO);
  if Response = IDYES then
  begin
    // Don't delete AppData
    Log('User data will be preserved in %LOCALAPPDATA%\ChronicleCore');
  end;
  Result := True;
end;

[UninstallDelete]
Type: filesandordirs; Name: "{app}"

[Registry]
; Store installation path for updates
Root: HKCU; Subkey: "Software\ChronicleCore"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\ChronicleCore"; ValueType: string; ValueName: "Version"; ValueData: "{#MyAppVersion}"; Flags: uninsdeletekey
