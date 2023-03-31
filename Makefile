clean:
	rm -rf build/bin

build: clean
	npm --prefix ./frontend run build && wails build -clean

launch: build
	open ./build/bin/Gvmapp.app/Contents/MacOS/Gvmapp

test:
	go test ./... -v

release:
	npm --prefix ./frontend run build && wails build -clean -nsis -webview2 download -platform darwin/universal,windows/amd64

sign: release
	gon -log-level=info ./build/darwin/gon-sign.json

notarize:
	xcrun altool --notarize-app --primary-bundle-id "com.uvite.Gvmapp" -u "pelotoken@gmail.com" -p "@env:APPLE_ID_PASSWORD" --asc-provider NA229UVJJB --file ./build/darwin/Gvmapp.dmg --output-format xml
