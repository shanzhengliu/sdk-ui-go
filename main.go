package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"sdk-ui-go/internal"
	"strings"
	"sync"
)

var (
	sdkmanInitScript = "~/.sdkman/bin/sdkman-init.sh"
	candidate        = make(map[string][]VersionMenu)
)

type VersionMenu struct {
	MenuItem *systray.MenuItem
	Title    string
}

func main() {
	systray.Run(OnReady, onExit)
}

func OnReady() {
	systray.SetIcon(internal.Icon)
	systray.SetTitle("SDK")
	systray.SetTooltip("SDK UI")
	internal.InstallSDKMan()
	internal.InstallNVM()
	candidate := internal.CandidateList(sdkmanInitScript)

	var wg sync.WaitGroup
	var candidateMenuItemMap = make(map[string]*systray.MenuItem)
	for _, c := range candidate {
		item := systray.AddMenuItem(c, "")
		candidateMenuItemMap[c] = item
	}

	for _, c := range candidate {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			item := candidateMenuItemMap[c]
			addSubMenu(item, c)
		}(c)
	}
	wg.Wait()
	systray.AddSeparator()

	nvmSubMenu()

	systray.AddSeparator()
	mSDKManVersion := systray.AddMenuItem("SDKMan Version", "")
	sdkmanUpdateItem := systray.AddMenuItem("SDKMan Update", "")
	systray.AddSeparator()
	nvmVersionItem := systray.AddMenuItem("NVM Version", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-sdkmanUpdateItem.ClickedCh:
				beeep.Notify("SDKMan Update", "SDKMan is updating", "")
				internal.SDKManUpdate(sdkmanInitScript)
				beeep.Notify("SDKMan Update", "SDKMan has updated", "")
			case <-mSDKManVersion.ClickedCh:
				beeep.Notify("SDKMan Version", internal.SDKManVersion(sdkmanInitScript), "")
			case <-nvmVersionItem.ClickedCh:
				beeep.Notify("NVM Version", internal.NVMVersion(), "")
			}

		}
	}()

}

func addSubMenu(item *systray.MenuItem, title string) {
	var versions []internal.Candidate
	var versionMenu []VersionMenu
	if strings.EqualFold(title, "Java") {
		versions = internal.JavaVersionList(sdkmanInitScript)
	} else {
		versions = internal.OtherVersionList(title, sdkmanInitScript)
	}
	versions = internal.SortCandidates(versions)
	addCustomItem := item.AddSubMenuItem("+ local "+title, "")
	go func() {
		for {
			select {
			case <-addCustomItem.ClickedCh:
				id := internal.AddCustomCandidate(title, sdkmanInitScript)
				if id != "" {
					customItem := item.AddSubMenuItem(id+"[Installed]", "")
					candidate[title] = append(candidate[title], VersionMenu{MenuItem: customItem, Title: title})
					addVersionItem(customItem, title, id, true)
				}
			}
		}
	}()

	for _, v := range versions {
		subItem := v.Identifier
		if v.Install {
			subItem = subItem + "[Installed]"
		} else {
			subItem = subItem + ""
		}

		versionItem := item.AddSubMenuItemCheckbox(subItem, "", v.Use)
		versionMenu = append(versionMenu, VersionMenu{MenuItem: versionItem, Title: title})
		addVersionItem(versionItem, title, v.Identifier, v.Install)
	}
	candidate[title] = versionMenu
}

func addVersionItem(item *systray.MenuItem, title string, version string, install bool) {
	installItem := item.AddSubMenuItem("Install && Use", "")
	uninstallItem := item.AddSubMenuItem("Uninstall", "")
	openHomeItem := item.AddSubMenuItem("Open Home", "")
	if install == false {
		uninstallItem.Hide()
		openHomeItem.Hide()
	}
	go func() {
		for {
			select {
			case <-installItem.ClickedCh:
				beeep.Notify("Install", "Verify Installation of "+title+" "+version, "")
				internal.UseCandidate(title, version, sdkmanInitScript)
				beeep.Notify("Install", title+" "+version+" has installed and Using", "")
				for _, v := range candidate[title] {
					if v.MenuItem != item {
						v.MenuItem.Uncheck()
					} else {
						item.Check()
						item.SetTitle(version + "[Installed]")
					}
				}
				openHomeItem.Show()
				uninstallItem.Show()
				installItem.Show()

			case <-uninstallItem.ClickedCh:
				if item.Checked() {
					return
				}
				beeep.Notify("Uninstall", "Uninstalling "+title+" "+version, "")
				internal.UninstallCandidate(title, version, sdkmanInitScript)
				beeep.Notify("Uninstall", title+" "+version+" has removed", "")
				item.SetTitle(version)
				uninstallItem.Hide()
				openHomeItem.Hide()
				installItem.Show()

			case <-openHomeItem.ClickedCh:
				internal.OpenCandidateFolder(title, version, sdkmanInitScript)

			}

		}
	}()
}

func nvmSubMenu() {
	nodeItem := systray.AddMenuItem("node", "")
	nodeVersionList := internal.NodeVersionList()
	var versionMenu []VersionMenu
	nodeVersionList = internal.SortCandidates(nodeVersionList)
	for _, v := range nodeVersionList {
		subItem := v.Identifier
		if v.Install {
			subItem = subItem + "[Installed]"
		} else {
			subItem = subItem + ""
		}
		versionItem := nodeItem.AddSubMenuItemCheckbox(subItem, "", v.Use)
		versionMenu = append(versionMenu, VersionMenu{MenuItem: versionItem, Title: "node[nvm]"})
		AddNodeVersionItem(versionItem, "node[nvm]", v.Identifier, v.Install)
	}
	candidate["node[nvm]"] = versionMenu

}

func AddNodeVersionItem(item *systray.MenuItem, title string, version string, install bool) {
	installItem := item.AddSubMenuItem("Install && Use", "")
	uninstallItem := item.AddSubMenuItem("Uninstall", "")
	openHomeItem := item.AddSubMenuItem("Open Home", "")
	if install == false {
		uninstallItem.Hide()
		openHomeItem.Hide()
	}
	go func() {
		for {
			select {
			case <-installItem.ClickedCh:
				beeep.Notify("Install", "Verify Installation of "+title+" "+version, "")
				internal.InstallNode(version)
				beeep.Notify("Install", title+" "+version+" has installed and Using", "")
				for _, v := range candidate["node[nvm]"] {
					if v.MenuItem != item {
						v.MenuItem.Uncheck()
					} else {
						item.Check()
						item.SetTitle(version + "[Installed]")
					}
				}
				openHomeItem.Show()
				uninstallItem.Show()
				installItem.Show()

			case <-uninstallItem.ClickedCh:
				if item.Checked() {
					return
				}
				beeep.Notify("Uninstall", "Uninstalling "+title+" "+version, "")
				internal.UninstallNode(version)
				beeep.Notify("Uninstall", title+" "+version+" has removed", "")
				item.SetTitle(version)
				uninstallItem.Hide()
				openHomeItem.Hide()
				installItem.Show()

			case <-openHomeItem.ClickedCh:
				internal.OpenNodeFolder(version)
			}
		}
	}()
}

func onExit() {
	// clean up here
	fmt.Println("Exiting...")
}
