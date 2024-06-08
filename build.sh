#!/bin/bash

APP_NAME="SDKUI"
EXECUTABLE_NAME="sdkuigo"
IDENTIFIER="com.shanzhengliu.sdkuigo"

# 编译 Go 应用程序
go build -o ${EXECUTABLE_NAME} main.go

# 创建目录结构
mkdir -p ${APP_NAME}.app/Contents/MacOS
mkdir -p ${APP_NAME}.app/Contents/Resources

# 移动文件
mv ${EXECUTABLE_NAME} ${APP_NAME}.app/Contents/MacOS/

# 创建 Info.plist 文件
cat <<EOF > ${APP_NAME}.app/Contents/Info.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>${EXECUTABLE_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>${IDENTIFIER}</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
</dict>
</plist>
EOF

echo "App bundle ${APP_NAME}.app created successfully."
