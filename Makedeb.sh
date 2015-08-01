

fakeroot fpm -C "rootfs" \
             -n "northstar" -v 1`git rev-parse HEAD | cut -c 1-12` \
             -p "northstar_1`git rev-parse HEAD | cut -c 1-12`_amd64.deb" \
             -s "dir" -t "deb" "usr"
