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
PKGFOLDER=$ENV_TAG-$PKGARCH-$PKGVERSION/mesh
PKGFILE=mesh-$ENV_TAG-$PKGARCH-$PKGVERSION.tar.gz
PKGREPLACES=mesh

if [ $PKGBRANCH = "master" ]; then
  PKGREPLACES=mesh-develop
fi

if [ $PKGARCH = "armv7" ]; then GOOS=linux GOARCH=arm GOARM=7 ./build
else
  echo "Specify PKGARCH=armv7"
  exit 1
fi

echo "Building $PKGFOLDER"

rm -rf /tmp/$PKGFOLDER

mkdir -p /tmp/$PKGFOLDER/
mkdir -p /tmp/$PKGFOLDER/log
mkdir -p /tmp/$PKGFOLDER/tmp
chmod 0775 /tmp/$PKGFOLDER/ -R

echo "coping ui package..."
cp contrib/ui/nas-drobo/Content/* /tmp/$PKGFOLDER/ -r
cp contrib/ui/www/* /tmp/$PKGFOLDER/www/ -r

cat > /tmp/$PKGFOLDER/version.txt << EOF
$PKGVERSION
EOF

cp mesh /tmp/$PKGFOLDER/
cp meshctl /tmp/$PKGFOLDER/
cp LICENSE /tmp/$PKGFOLDER/
chmod +x /tmp/$PKGFOLDER/mesh /tmp/$PKGFOLDER/meshctl
chmod +x /tmp/$PKGFOLDER/*.sh
chmod 0775 /tmp/$PKGFOLDER/www -R

cd /tmp/$PKGFOLDER && tar czf ../mesh.tgz $(ls .)
cd ../ && md5sum mesh.tgz > mesh.tgz.md5
tar czf $PKGFILE mesh.tgz mesh.tgz.md5

#rm -rf /tmp/$PKGFOLDER/
#mv *.apk $PKGFILE
