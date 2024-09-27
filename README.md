# ⚠️ Warning ⚠️

You use it at your own risk. Overheating can destroy your hardware.

# 💡 About

That's a small service to control fan speed. It reads temperature sensors and current power profile and maps them to fan level. Mapping is set by the configuration file. There is no user interface, and there is no plan to make one.

Main features:

* Multiple fans and sensors.
* Power profile support.
* Delay before changing fan speed.

Originally created for Thinkpad T14 gen 4 (Intel). There is [thinkfan](https://github.com/vmatare/thinkfan), but it does not support power profiles.

There are other fan [speed control tools](https://wiki.archlinux.org/title/fan_speed_control).

Adding new drivers and other improvements are welcome.

# ⚙️ Supported drivers

For now it support only thinkpads, but it's not that hard to add new drivers.

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

It should work on every device, but you need to find sensor name and label.

```bash
tail -n 1 $(ls /sys/class/hwmon/hwmon*/{name,temp*_label,temp*_input} | sort)
```

#### Links

* [Hwmon kernel documentation](https://www.kernel.org/doc/Documentation/hwmon/sysfs-interface)

## 🚀 Profile platform

There is a file `/sys/firmware/acpi/platform_profile` that contains current power profile. In KDE and GNOME you can control current profile from the user interface.

```bash
# To see available profiles
cat /sys/firmware/acpi/platform_profile_choices
```

#### Links

* [Kernel commit](https://patchwork.kernel.org/project/linux-acpi/patch/20201218174759.667457-2-markpearson@lenovo.com/)

# 🧪 Configuration

There is [conf/fanctl.yaml](conf/fanctl.yaml) file with all available parameters and some explanation. You can use it to create your own config.

[conf/thinkpad.yaml](conf/thinkpad.yaml) is an example I tested on T14.

Binary file has `-d` flag which enables debug logs. And `-c` flag to pass the config file path.

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
sudo systemctl enable fanctl.service fanctl-wakeup.service fanctl-suspend.service
sudo systemctl start fanctl
# Check service status
sudo systemctl status fanctl
```

## Deb package

Download deb package [here](https://github.com/IvanSafonov/fanctl/releases) and install. 

```bash
sudo apt install ./fanctl*.deb
# Change /etc/fanctl.yaml according to your hardware
sudo systemctl enable fanctl.service fanctl-wakeup.service fanctl-suspend.service
sudo systemctl start fanctl
# Check service status
sudo systemctl status fanctl
```
