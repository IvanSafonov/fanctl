# ⚠️ Warning ⚠️

Think twice before use it. You use on your own risk.
Overheating can destroy your hardware.

# 💡 About

That's a small service to control fan speed. It reads temperature sensors and current power profile and maps it to fan level. Mapping is set by configuration file. There is no user interface and there is no plan to make it.

Main features:

* Multiple fans and sensors.
* Power profile support.
* Delay before changing fan speed.

Originally created for Thinkpad T14 gen 4 (Intel). There is thinkfan but it does not support power profiles.

There are other fan [speed control tools](https://wiki.archlinux.org/title/fan_speed_control).

Adding new drivers and other improvements are welcome.

# ⚙️ Supported drivers

## 💻  Thinkpad ACPI Fan

❗Fan control is disabled by default. You need to enable it.

```bash
echo "options thinkpad_acpi fan_control=1" | sudo tee -a /usr/lib/modprobe.d/thinkpad_acpi.conf
sudo modprobe thinkpad_acpi
```

#### Links

* [Arch wiki](https://wiki.archlinux.org/title/fan_speed_control#ThinkPad_laptops)
* [Thinkpad acpi documentation](https://www.kernel.org/doc/Documentation/laptops/thinkpad-acpi.txt)

## 🌡️ Hwmon sensors

It should work on every device but you need to find sensor name and label.

```bash
tail -n 1 $(ls /sys/class/hwmon/hwmon*/{name,temp*_label,temp*_input} | sort)
```

#### Links

* [Hwmon kernel documentation](https://www.kernel.org/doc/Documentation/hwmon/sysfs-interface)

## 🚀 Profile platform

There is a file `/sys/firmware/acpi/platform_profile` which contains current power profile. In KDE and GNOME you can control current profile from the user interface.

```bash
# To see available profiles
cat /sys/firmware/acpi/platform_profile_choices
```

#### Links

* [Kernel commit](https://patchwork.kernel.org/project/linux-acpi/patch/20201218174759.667457-2-markpearson@lenovo.com/)

# 🧪 Testing config

Binary file has `-d` flag which enables debug logs. And `-c` flag to pass config file path.

```bash
sudo fanctl -d -c ./conf/fanctl.yaml
```

# 📦 Install

## Manual

Install [latest version of go](https://go.dev/doc/install)

```bash
cd ./fanctl
go build -o fanctl ./cmd/fanctl
sudo cp ./fanctl /usr/sbin/fanctl
sudo cp ./systemd/* /lib/systemd/system/
sudo cp ./conf/fanctl.yaml /etc/
# Change /etc/fanctl.yaml according to your hardware
sudo systemctl enable fanctl.service fanctl-wakeup.service
sudo systemctl start fanctl
# Check service status
sudo systemctl status fanctl
```

## Deb package

Run `build-deb.sh` to build deb package.

```bash
./build-deb.sh
sudo apt install ./fanctl*.deb
# Change /etc/fanctl.yaml according to your hardware
sudo systemctl enable fanctl.service fanctl-wakeup.service
sudo systemctl start fanctl
# Check service status
sudo systemctl status fanctl
```
