#!/bin/sh

# This script generates an MSI file for RiV-mesh for a given architecture. It
# needs to run on Windows within MSYS2 and Go 1.13 or later must be installed on
# the system and within the PATH. This is ran currently by Appveyor (see
# appveyor.yml in the repository root) for both x86 and x64.
#
# Author: Neil Alexander <neilalexander@users.noreply.github.com>

# Get arch from command line if given
PKGARCH="x64"
#if [ "${PKGARCH}" == "" ];
#then
#  echo "tell me the architecture: x86, x64 or arm"
#  exit 1
#fi

# Get the rest of the repository history. This is needed within Appveyor because
# otherwise we don't get all of the branch histories and therefore the semver
# scripts don't work properly.
if [ "${APPVEYOR_PULL_REQUEST_HEAD_REPO_BRANCH}" != "" ];
then
  git fetch --all
  git checkout ${APPVEYOR_PULL_REQUEST_HEAD_REPO_BRANCH}
elif [ "${APPVEYOR_REPO_BRANCH}" != "" ];
then
  git fetch --all
  git checkout ${APPVEYOR_REPO_BRANCH}
fi

# Install prerequisites within MSYS2
pacman -S --needed --noconfirm unzip git curl

# Download the wix tools!
if [ ! -d wixbin ];
then
  curl -LO https://github.com/wixtoolset/wix3/releases/download/wix3112rtm/wix311-binaries.zip
  if [ `md5sum wix311-binaries.zip | cut -f 1 -d " "` != "47a506f8ab6666ee3cc502fb07d0ee2a" ];
  then
    echo "wix package didn't match expected checksum"
    exit 1
  fi
  mkdir -p wixbin
  unzip -o wix311-binaries.zip -d wixbin || (
    echo "failed to unzip WiX"
    exit 1
  )
fi

# Build RiV-mesh!
[ "${PKGARCH}" == "x64" ] && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 ./build
[ "${PKGARCH}" == "x86" ] && GOOS=windows GOARCH=386 CGO_ENABLED=0 ./build
[ "${PKGARCH}" == "arm" ] && GOOS=windows GOARCH=arm CGO_ENABLED=0 ./build
#[ "${PKGARCH}" == "arm64" ] && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 ./build

# Create the postinstall script
cat > updateconfig.bat << EOF
if not exist %ALLUSERSPROFILE%\\RiV-mesh (
  mkdir %ALLUSERSPROFILE%\\RiV-mesh
)
if not exist %ALLUSERSPROFILE%\\RiV-mesh\\mesh.conf (
  if exist mesh.exe (
    mesh.exe -genconf > %ALLUSERSPROFILE%\\RiV-mesh\\mesh.conf
  )
)
EOF

# Work out metadata for the package info
PKGNAME=$(sh contrib/semver/name.sh)
PKGVERSION=$(sh contrib/msi/msversion.sh --bare)
PKGVERSIONMS=$(echo $PKGVERSION | tr - .)
[ "${PKGARCH}" == "x64" ] && \
  PKGGUID="77757838-1a23-40a5-a720-c3b43e0260cc" PKGINSTFOLDER="ProgramFiles64Folder" || \
  PKGGUID="54a3294e-a441-4322-aefb-3bb40dd022bb" PKGINSTFOLDER="ProgramFilesFolder"

# Download the Wintun driver
if [ ! -d wintun ];
then
  curl -o wintun.zip https://www.wintun.net/builds/wintun-0.11.zip
  unzip wintun.zip
fi
if [ $PKGARCH = "x64" ]; then
  PKGWINTUNDLL=wintun/bin/amd64/wintun.dll
elif [ $PKGARCH = "x86" ]; then
  PKGWINTUNDLL=wintun/bin/x86/wintun.dll
elif [ $PKGARCH = "arm" ]; then
  PKGWINTUNDLL=wintun/bin/arm/wintun.dll
#elif [ $PKGARCH = "arm64" ]; then
#  PKGWINTUNDLL=wintun/bin/arm64/wintun.dll
else
  echo "wasn't sure which architecture to get wintun for"
  exit 1
fi

if [ $PKGNAME != "master" ]; then
  PKGDISPLAYNAME="RiV-mesh Network (${PKGNAME} branch)"
else
  PKGDISPLAYNAME="RiV-mesh Network"
fi

# Generate the wix.xml file
cat > wix.xml << EOF
<?xml version="1.0" encoding="windows-1252"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product
    Name="${PKGDISPLAYNAME}"
    Id="*"
    UpgradeCode="${PKGGUID}"
    Language="1033"
    Codepage="1252"
    Version="${PKGVERSIONMS}"
    Manufacturer="github.com/RiV-chain">

    <Package
      Id="*"
      Keywords="Installer"
      Description="RiV-mesh Network Installer"
      Comments="RiV-mesh Network standalone router for Windows."
      Manufacturer="github.com/RiV-chain"
      InstallerVersion="200"
      InstallScope="perMachine"
      Languages="1033"
      Compressed="yes"
      SummaryCodepage="1252" />

    <MajorUpgrade
      AllowDowngrades="yes" />

    <Media
      Id="1"
      Cabinet="Media.cab"
      EmbedCab="yes"
      CompressionLevel="high" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="${PKGINSTFOLDER}" Name="PFiles">
        <Directory Id="RiVMeshInstallFolder" Name="RiV-mesh">

          <Component Id="MainExecutable" Guid="ca425a1c-407b-4d37-a629-6d004922e5e6">
            <File
              Id="RiVMesh"
              Name="mesh.exe"
              DiskId="1"
              Source="mesh.exe"
              KeyPath="yes" />

            <File
              Id="Wintun"
              Name="wintun.dll"
              DiskId="1"
              Source="${PKGWINTUNDLL}" />

            <ServiceInstall
              Id="ServiceInstaller"
              Account="LocalSystem"
              Description="RiV-mesh Network router process"
              DisplayName="RiV-mesh Service"
              ErrorControl="normal"
              LoadOrderGroup="NetworkProvider"
              Name="RiV-mesh"
              Start="auto"
              Type="ownProcess"
              Arguments='-useconffile "%ALLUSERSPROFILE%\\RiV-mesh\\mesh.conf" -logto "%ALLUSERSPROFILE%\\RiV-mesh\\mesh.log"'
              Vital="yes" />

            <ServiceControl
              Id="ServiceControl"
              Name="RiV-mesh"
              Start="install"
              Stop="both"
              Remove="uninstall" />
          </Component>

          <Component Id="CtrlExecutable" Guid="7c13026e-8654-4ccb-8ac3-7cc11f284e52">
            <File
              Id="RiVMeshctl"
              Name="meshctl.exe"
              DiskId="1"
              Source="meshctl.exe"
              KeyPath="yes"/>
          </Component>

          <Component Id="ConfigScript" Guid="b9e3d00c-25f5-413c-b152-f8def1cf5025">
            <File
              Id="Configbat"
              Name="updateconfig.bat"
              DiskId="1"
              Source="updateconfig.bat"
              KeyPath="yes"/>
          </Component>
        </Directory>
      </Directory>
    </Directory>

    <Feature Id="RiVMeshFeature" Title="RiV-mesh" Level="1">
      <ComponentRef Id="MainExecutable" />
      <ComponentRef Id="CtrlExecutable" />
      <ComponentRef Id="ConfigScript" />
    </Feature>

    <CustomAction
      Id="UpdateGenerateConfig"
      Directory="RiVMeshInstallFolder"
      ExeCommand="cmd.exe /c updateconfig.bat"
      Execute="deferred"
      Return="check"
      Impersonate="yes" />

    <InstallExecuteSequence>
      <Custom
        Action="UpdateGenerateConfig"
        Before="StartServices">
          NOT Installed AND NOT REMOVE
      </Custom>
    </InstallExecuteSequence>

  </Product>
</Wix>
EOF

# Generate the MSI
CANDLEFLAGS="-nologo"
LIGHTFLAGS="-nologo -spdb -sice:ICE71 -sice:ICE61"
wixbin/candle.exe $CANDLEFLAGS -out ${PKGNAME}-${PKGVERSION}-${PKGARCH}.wixobj -arch ${PKGARCH} wix.xml && \
wixbin/light.exe $LIGHTFLAGS -ext WixUtilExtension.dll -out ${PKGNAME}-${PKGVERSION}-${PKGARCH}.msi ${PKGNAME}-${PKGVERSION}-${PKGARCH}.wixobj
