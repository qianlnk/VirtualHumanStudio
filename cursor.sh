#!/bin/bash

new_machine_id=$(uuidgen | tr '[:upper:]' '[:lower:]')
new_dev_device_id=$(uuidgen | tr '[:upper:]' '[:lower:]')
new_mac_machine_id=$(openssl rand -hex 32)

echo $new_machine_id > ~/Library/Application\ Support/Cursor/machineid
sed -i '' "s/\"telemetry.devDeviceId\": \".*\"/\"telemetry.devDeviceId\": \"$new_dev_device_id\"/" ~/Library/Application\ Support/Cursor/User/globalStorage/storage.json
sed -i '' "s/\"telemetry.macMachineId\": \".*\"/\"telemetry.macMachineId\": \"$new_mac_machine_id\"/" ~/Library/Application\ Support/Cursor/User/globalStorage/storage.json