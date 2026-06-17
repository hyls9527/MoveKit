package folders

type KnownFolder struct {
	Name         string // "桌面"
	GUID         string // "{B4BFCC3A-DB2C-424C-B029-7FE99A87C641}"
	RegValueName string // "Desktop" (Shell Folders key name)
}

var AllFolders = []KnownFolder{
	{Name: "桌面", GUID: "{B4BFCC3A-DB2C-424C-B029-7FE99A87C641}", RegValueName: "Desktop"},
	{Name: "下载", GUID: "{374DE290-123F-4565-9164-39C4925E467B}", RegValueName: "{374DE290-123F-4565-9164-39C4925E467B}"},
	{Name: "音乐", GUID: "{4BD8D571-6D19-4D8D-BE92-CC209D310997}", RegValueName: "My Music"},
	{Name: "文档", GUID: "{FDD39AD0-238F-46AF-ADB4-6C85480369C7}", RegValueName: "Personal"},
	{Name: "图片", GUID: "{33E28130-4E1E-4676-835A-98395C3BC3BB}", RegValueName: "My Pictures"},
}

var GUIDToRegValue = map[string]string{
	"{B4BFCC3A-DB2C-424C-B029-7FE99A87C641}": "Desktop",
	"{374DE290-123F-4565-9164-39C4925E467B}": "{374DE290-123F-4565-9164-39C4925E467B}",
	"{4BD8D571-6D19-4D8D-BE92-CC209D310997}": "My Music",
	"{FDD39AD0-238F-46AF-ADB4-6C85480369C7}": "Personal",
	"{33E28130-4E1E-4676-835A-98395C3BC3BB}": "My Pictures",
}
