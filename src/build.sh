# Syncthing Specific
cd ..
find -type f -name '*sync-conflict*' -delete
cd src

go-bindata-assetfs -debug -o bindata_assetfs.go resources/app/...
mv bindata_assetfs.go backend/bins.go
go build -o OttoDJ backend/main.go backend/bins.go backend/fader.go
GOOS=linux gopherjs build client/client.go -o resources/app/scripts/client.js
