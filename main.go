package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
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
	systray.SetIcon(icon.Data)
	systray.SetTitle("SDK UI")
	systray.SetTooltip("SDK UI")
	internal.InstallSDKMan()
	candidate := internal.CandidateList(sdkmanInitScript)

	var wg sync.WaitGroup
	var candidateMenuItemMap = make(map[string]*systray.MenuItem)
	for _, c := range candidate {
		item := systray.AddMenuItem(c, c)
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
	mSDKManVersion := systray.AddMenuItem("SDKMan Version", "Version")
	sdkmanUpdateItem := systray.AddMenuItem("SDKMan Update", "Update")
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
				beeep.Notify("SDKMan Version", internal.SDKManVersion(sdkmanInitScript), "SDKMan Version")
			}
		}
	}()

}

func onExit() {
	// Clean up here
	fmt.Println("Exiting")
}

func addSubMenu(item *systray.MenuItem, title string) {
	var versions []internal.Candidate
	var versionMenu []VersionMenu
	if strings.EqualFold(title, "Java") {
		versions = internal.JavaVersionList(sdkmanInitScript)
	} else {
		versions = internal.OtherVersionList(title, sdkmanInitScript)
	}

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
