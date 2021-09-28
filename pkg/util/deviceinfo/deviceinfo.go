package deviceinfo

import (
	"fmt"
)

// Credit to https://github.com/react-native-device-info/react-native-device-info/blob/master/ios/RNDeviceInfo/RNDeviceInfo.m
var iosMachineMapping = map[string]string{
	"iPod1,1":    "iPod Touch",     // (Original)
	"iPod2,1":    "iPod Touch",     // (Second Generation)
	"iPod3,1":    "iPod Touch",     // (Third Generation)
	"iPod4,1":    "iPod Touch",     // (Fourth Generation)
	"iPod5,1":    "iPod Touch",     // (Fifth Generation)
	"iPod7,1":    "iPod Touch",     // (Sixth Generation)
	"iPod9,1":    "iPod Touch",     // (Seventh Generation)
	"iPhone1,1":  "iPhone",         // (Original)
	"iPhone1,2":  "iPhone 3G",      // (3G)
	"iPhone2,1":  "iPhone 3GS",     // (3GS)
	"iPad1,1":    "iPad",           // (Original)
	"iPad2,1":    "iPad 2",         //
	"iPad2,2":    "iPad 2",         //
	"iPad2,3":    "iPad 2",         //
	"iPad2,4":    "iPad 2",         //
	"iPad3,1":    "iPad",           // (3rd Generation)
	"iPad3,2":    "iPad",           // (3rd Generation)
	"iPad3,3":    "iPad",           // (3rd Generation)
	"iPhone3,1":  "iPhone 4",       // (GSM)
	"iPhone3,2":  "iPhone 4",       // iPhone 4
	"iPhone3,3":  "iPhone 4",       // (CDMA/Verizon/Sprint)
	"iPhone4,1":  "iPhone 4S",      //
	"iPhone5,1":  "iPhone 5",       // (model A1428, AT&T/Canada)
	"iPhone5,2":  "iPhone 5",       // (model A1429, everything else)
	"iPad3,4":    "iPad",           // (4th Generation)
	"iPad3,5":    "iPad",           // (4th Generation)
	"iPad3,6":    "iPad",           // (4th Generation)
	"iPad2,5":    "iPad Mini",      // (Original)
	"iPad2,6":    "iPad Mini",      // (Original)
	"iPad2,7":    "iPad Mini",      // (Original)
	"iPhone5,3":  "iPhone 5c",      // (model A1456, A1532 | GSM)
	"iPhone5,4":  "iPhone 5c",      // (model A1507, A1516, A1526 (China), A1529 | Global)
	"iPhone6,1":  "iPhone 5s",      // (model A1433, A1533 | GSM)
	"iPhone6,2":  "iPhone 5s",      // (model A1457, A1518, A1528 (China), A1530 | Global)
	"iPhone7,1":  "iPhone 6 Plus",  //
	"iPhone7,2":  "iPhone 6",       //
	"iPhone8,1":  "iPhone 6s",      //
	"iPhone8,2":  "iPhone 6s Plus", //
	"iPhone8,4":  "iPhone SE",      //
	"iPhone9,1":  "iPhone 7",       // (model A1660 | CDMA)
	"iPhone9,3":  "iPhone 7",       // (model A1778 | Global)
	"iPhone9,2":  "iPhone 7 Plus",  // (model A1661 | CDMA)
	"iPhone9,4":  "iPhone 7 Plus",  // (model A1784 | Global)
	"iPhone10,3": "iPhone X",       // (model A1865, A1902)
	"iPhone10,6": "iPhone X",       // (model A1901)
	"iPhone10,1": "iPhone 8",       // (model A1863, A1906, A1907)
	"iPhone10,4": "iPhone 8",       // (model A1905)
	"iPhone10,2": "iPhone 8 Plus",  // (model A1864, A1898, A1899)
	"iPhone10,5": "iPhone 8 Plus",  // (model A1897)
	"iPhone11,2": "iPhone XS",      // (model A2097, A2098)
	"iPhone11,4": "iPhone XS Max",  // (model A1921, A2103)
	"iPhone11,6": "iPhone XS Max",  // (model A2104)
	"iPhone11,8": "iPhone XR",      // (model A1882, A1719, A2105)
	"iPhone12,1": "iPhone 11",
	"iPhone12,3": "iPhone 11 Pro",
	"iPhone12,5": "iPhone 11 Pro Max",
	"iPhone13,1": "iPhone 12 mini",
	"iPhone13,2": "iPhone 12",
	"iPhone13,3": "iPhone 12 Pro",
	"iPhone13,4": "iPhone 12 Pro Max",
	"iPhone12,8": "iPhone SE",                           // (2nd Generation iPhone SE)
	"iPad4,1":    "iPad Air",                            // 5th Generation iPad (iPad Air) - Wifi
	"iPad4,2":    "iPad Air",                            // 5th Generation iPad (iPad Air) - Cellular
	"iPad4,3":    "iPad Air",                            // 5th Generation iPad (iPad Air)
	"iPad4,4":    "iPad Mini 2",                         // (2nd Generation iPad Mini - Wifi)
	"iPad4,5":    "iPad Mini 2",                         // (2nd Generation iPad Mini - Cellular)
	"iPad4,6":    "iPad Mini 2",                         // (2nd Generation iPad Mini)
	"iPad4,7":    "iPad Mini 3",                         // (3rd Generation iPad Mini)
	"iPad4,8":    "iPad Mini 3",                         // (3rd Generation iPad Mini)
	"iPad4,9":    "iPad Mini 3",                         // (3rd Generation iPad Mini)
	"iPad5,1":    "iPad Mini 4",                         // (4th Generation iPad Mini)
	"iPad5,2":    "iPad Mini 4",                         // (4th Generation iPad Mini)
	"iPad5,3":    "iPad Air 2",                          // 6th Generation iPad (iPad Air 2)
	"iPad5,4":    "iPad Air 2",                          // 6th Generation iPad (iPad Air 2)
	"iPad6,3":    "iPad Pro 9.7-inch",                   // iPad Pro 9.7-inch
	"iPad6,4":    "iPad Pro 9.7-inch",                   // iPad Pro 9.7-inch
	"iPad6,7":    "iPad Pro 12.9-inch",                  // iPad Pro 12.9-inch
	"iPad6,8":    "iPad Pro 12.9-inch",                  // iPad Pro 12.9-inch
	"iPad6,11":   "iPad (5th generation)",               // Apple iPad 9.7 inch (5th generation) - WiFi
	"iPad6,12":   "iPad (5th generation)",               // Apple iPad 9.7 inch (5th generation) - WiFi + cellular
	"iPad7,1":    "iPad Pro 12.9-inch",                  // 2nd Generation iPad Pro 12.5-inch - Wifi
	"iPad7,2":    "iPad Pro 12.9-inch",                  // 2nd Generation iPad Pro 12.5-inch - Cellular
	"iPad7,3":    "iPad Pro 10.5-inch",                  // iPad Pro 10.5-inch - Wifi
	"iPad7,4":    "iPad Pro 10.5-inch",                  // iPad Pro 10.5-inch - Cellular
	"iPad7,5":    "iPad (6th generation)",               // iPad (6th generation) - Wifi
	"iPad7,6":    "iPad (6th generation)",               // iPad (6th generation) - Cellular
	"iPad7,11":   "iPad (7th generation)",               // iPad 10.2 inch (7th generation) - Wifi
	"iPad7,12":   "iPad (7th generation)",               // iPad 10.2 inch (7th generation) - Wifi + cellular
	"iPad8,1":    "iPad Pro 11-inch (3rd generation)",   // iPad Pro 11 inch (3rd generation) - Wifi
	"iPad8,2":    "iPad Pro 11-inch (3rd generation)",   // iPad Pro 11 inch (3rd generation) - 1TB - Wifi
	"iPad8,3":    "iPad Pro 11-inch (3rd generation)",   // iPad Pro 11 inch (3rd generation) - Wifi + cellular
	"iPad8,4":    "iPad Pro 11-inch (3rd generation)",   // iPad Pro 11 inch (3rd generation) - 1TB - Wifi + cellular
	"iPad8,5":    "iPad Pro 12.9-inch (3rd generation)", // iPad Pro 12.9 inch (3rd generation) - Wifi
	"iPad8,6":    "iPad Pro 12.9-inch (3rd generation)", // iPad Pro 12.9 inch (3rd generation) - 1TB - Wifi
	"iPad8,7":    "iPad Pro 12.9-inch (3rd generation)", // iPad Pro 12.9 inch (3rd generation) - Wifi + cellular
	"iPad8,8":    "iPad Pro 12.9-inch (3rd generation)", // iPad Pro 12.9 inch (3rd generation) - 1TB - Wifi + cellular
	"iPad11,1":   "iPad Mini 5",                         // (5th Generation iPad Mini)
	"iPad11,2":   "iPad Mini 5",                         // (5th Generation iPad Mini)
	"iPad11,3":   "iPad Air (3rd generation)",
	"iPad11,4":   "iPad Air (3rd generation)",
	"AppleTV2,1": "Apple TV",    // Apple TV (2nd Generation)
	"AppleTV3,1": "Apple TV",    // Apple TV (3rd Generation)
	"AppleTV3,2": "Apple TV",    // Apple TV (3rd Generation - Rev A)
	"AppleTV5,3": "Apple TV",    // Apple TV (4th Generation)
	"AppleTV6,2": "Apple TV 4K", // Apple TV 4K
}

func DeviceModel(deviceInfo map[string]interface{}) string {
	android, ok := deviceInfo["android"].(map[string]interface{})
	if ok {
		return DeviceModelAndroid(android)
	}
	ios, ok := deviceInfo["ios"].(map[string]interface{})
	if ok {
		return DeviceModelIOS(ios)
	}
	return ""
}

func DeviceModelAndroid(android map[string]interface{}) string {
	build, ok := android["Build"].(map[string]interface{})
	if !ok {
		return ""
	}
	manufacturer, ok := build["MANUFACTURER"].(string)
	if !ok {
		return ""
	}
	model, ok := build["MODEL"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v %v", manufacturer, model)
}

func DeviceModelIOS(ios map[string]interface{}) string {
	uname, ok := ios["uname"].(map[string]interface{})
	if !ok {
		return ""
	}
	machine, ok := uname["machine"].(string)
	if !ok {
		return ""
	}

	// Simulator on non Apple Silicon macOS.
	if machine == "x86_64" {
		return "iOS Simulator"
	}

	// Simulator on Apple Silicon macOS.
	if machine == "arm64" {
		return "iOS Simulator"
	}

	known, ok := iosMachineMapping[machine]
	if ok {
		return known
	}

	return machine
}

func DeviceName(deviceInfo map[string]interface{}) string {
	android, ok := deviceInfo["android"].(map[string]interface{})
	if ok {
		return DeviceNameAndroid(android)
	}
	ios, ok := deviceInfo["ios"].(map[string]interface{})
	if ok {
		return DeviceNameIOS(ios)
	}
	return ""
}

func DeviceNameAndroid(android map[string]interface{}) string {
	if settings, ok := android["Settings"].(map[string]interface{}); ok {
		if global, ok := settings["Global"].(map[string]interface{}); ok {
			if deviceName, ok := global["DEVICE_NAME"].(string); ok {
				return deviceName
			}
		}
		if secure, ok := settings["Secure"].(map[string]interface{}); ok {
			if bluetoothName, ok := secure["bluetooth_name"].(string); ok {
				return bluetoothName
			}
		}
	}
	return ""
}

func DeviceNameIOS(ios map[string]interface{}) string {
	if uiDevice, ok := ios["UIDevice"].(map[string]interface{}); ok {
		if name, ok := uiDevice["name"].(string); ok {
			return name
		}
	}
	if uname, ok := ios["uname"].(map[string]interface{}); ok {
		if nodename, ok := uname["nodename"].(string); ok {
			return nodename
		}
	}
	return ""
}

func ApplicationName(deviceInfo map[string]interface{}) string {
	android, ok := deviceInfo["android"].(map[string]interface{})
	if ok {
		return ApplicationNameAndroid(android)
	}
	ios, ok := deviceInfo["ios"].(map[string]interface{})
	if ok {
		return ApplicationNameIOS(ios)
	}
	return ""
}

func ApplicationNameAndroid(android map[string]interface{}) string {
	if applicationInfoLabel, ok := android["ApplicationInfoLabel"].(string); ok {
		return applicationInfoLabel
	}
	return ""
}

func ApplicationNameIOS(ios map[string]interface{}) string {
	if nsBundle, ok := ios["NSBundle"].(map[string]interface{}); ok {
		if cfBundleDisplayName, ok := nsBundle["CFBundleDisplayName"].(string); ok {
			return cfBundleDisplayName
		}
	}
	return ""
}
