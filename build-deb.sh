set -e -x

NAME=fanctl
VERSION=0.1.1

TMP_PACKAGE_DIR=.debian
TMP_PACKAGE_DEB_DIR=$TMP_PACKAGE_DIR/DEBIAN

rm -rf $TMP_PACKAGE_DIR
mkdir -p $TMP_PACKAGE_DEB_DIR

CONTROL=$TMP_PACKAGE_DEB_DIR/control
CONFFILES=$TMP_PACKAGE_DEB_DIR/conffiles

ARCH=$(dpkg-architecture -qDEB_BUILD_ARCH)
PACKAGE_FILENAME=${NAME}_${VERSION}_${ARCH}.deb

echo "Package: $NAME" >> $CONTROL
echo "Version: $VERSION" >> $CONTROL
echo "Priority: optional" >> $CONTROL
echo "Architecture: $ARCH" >> $CONTROL
echo "Maintainer: Ivan Safonov <safonov.ivan.s@gmail.com>" >> $CONTROL
echo "Description: Fan control service" >> $CONTROL
echo "Homepage: https://github.com/IvanSafonov/fanctl" >> $CONTROL

cp ./debian/* $TMP_PACKAGE_DEB_DIR/

mkdir -p $TMP_PACKAGE_DIR/usr/sbin
go build -o $TMP_PACKAGE_DIR/usr/sbin/fanctl ./cmd/fanctl
strip $TMP_PACKAGE_DIR/usr/sbin/fanctl

mkdir -p $TMP_PACKAGE_DIR/etc
cp ./conf/fanctl.yaml $TMP_PACKAGE_DIR/etc/fanctl.yaml
echo "/etc/fanctl.yaml" >> $CONFFILES

mkdir -p $TMP_PACKAGE_DIR/lib/systemd/system
cp ./systemd/*.service $TMP_PACKAGE_DIR/lib/systemd/system/

mkdir -p $TMP_PACKAGE_DIR/usr/local/share/doc/fanctl/examples
cp ./conf/*.yaml $TMP_PACKAGE_DIR/usr/local/share/doc/fanctl/examples/

echo -n "Installed-Size: " >> $CONTROL
du -sx --exclude DEBIAN $TMP_PACKAGE_DIR | grep -o -E ^[0-9]+ >> $CONTROL

dpkg-deb --build $TMP_PACKAGE_DIR $PACKAGE_FILENAME

rm -rf $TMP_PACKAGE_DIR
