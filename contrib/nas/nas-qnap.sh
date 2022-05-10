#!/bin/sh

# This is a lazy script to create a .bin for WD NAS build.
# You can give it the PKGARCH= argument
# i.e. PKGARCH=x86_64 contrib/nas/nas-asustor.sh

if [ `pwd` != `git rev-parse --show-toplevel` ]
then
  echo "You should run this script from the top-level directory of the git repo"
  exit 1
fi

PKGBRANCH=$(basename `git name-rev --name-only HEAD`)
PKG=$(sh contrib/semver/name.sh)
PKGVERSION=$(sh contrib/semver/version.sh --bare)
PKGARCH=${PKGARCH-amd64}
PKGFOLDER=$ENV_TAG-$PKGARCH-$PKGVERSION
PKGFILE=mesh-$PKGFOLDER.qpkg
PKGREPLACES=mesh

if [ $PKGBRANCH = "master" ]; then
  PKGREPLACES=mesh-develop
fi

if [ $PKGARCH = "x86-64" ]; then GOOS=linux GOARCH=amd64 ./build
elif [ $PKGARCH = "armv7" ]; then GOOS=linux GOARCH=arm GOARM=7 ./build
else
  echo "Specify PKGARCH=x86-64 or armv7"
  exit 1
fi

echo "Building $PKGFOLDER"

rm -rf /tmp/$PKGFOLDER

mkdir -p /tmp/$PKGFOLDER/icons
mkdir -p /tmp/$PKGFOLDER/au/bin
mkdir -p /tmp/$PKGFOLDER/au/tmp
mkdir -p /tmp/$PKGFOLDER/au/lib
mkdir -p /tmp/$PKGFOLDER/au/www
mkdir -p /tmp/$PKGFOLDER/au/var/log
mkdir -p /tmp/$PKGFOLDER/au/var/lib/bin
chmod 0775 /tmp/$PKGFOLDER/ -R

echo "coping ui package..."
cp contrib/ui/nas-qnap/package/* /tmp/$PKGFOLDER/ -r
cp contrib/ui/nas-qnap/au/* /tmp/$PKGFOLDER/au -r
cp contrib/ui/www/* /tmp/$PKGFOLDER/au/www/ -r

echo "Converting icon for: 64x64"
convert -colorspace sRGB ./riv.png -resize 64x64 /tmp/$PKGFOLDER/icons/mesh.png
echo "Converting icon for: 80x80"
convert -colorspace sRGB ./riv.png -resize 80x80 /tmp/$PKGFOLDER/icons/mesh_80.png
convert -colorspace sRGB ./riv.png -resize 64x64 /tmp/$PKGFOLDER/icons/mesh_gray.png

cat > /tmp/$PKGFOLDER/qdk.conf << EOF
QPKG_DISPLAY_NAME="RiV Mesh"
QPKG_NAME="mesh"
QPKG_VER="$PKGVERSION"
QPKG_AUTHOR="Riv Chain ltd"
QPKG_SUMMARY="RiV-mesh is an implementation of a fully end-to-end encrypted IPv6 network."
QPKG_RC_NUM="198"
QPKG_SERVICE_PROGRAM="mesh.sh"
QPKG_WEBUI="/mesh"
QPKG_WEB_PORT=
QPKG_LICENSE="LGPLv3"
QDK_BUILD_ARCH="$PKGARCH"
EOF


cp mesh /tmp/$PKGFOLDER/au/bin
cp meshctl /tmp/$PKGFOLDER/au/bin
chmod +x /tmp/$PKGFOLDER/au/bin/*
chmod 0775 /tmp/$PKGFOLDER/au/www -R

cd /tmp/$PKGFOLDER && qbuild --force-config -v

#rm -rf /tmp/$PKGFOLDER/
#mv *.apk $PKGFILE
