package main

import (
	"fmt"
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
	candidate := internal.CandidateList(sdkmanInitScript)
	var wg sync.WaitGroup

	for _, c := range candidate {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			item := systray.AddMenuItem(c, c)
			addSubMenu(item, c)
		}(c)
	}
	wg.Wait()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
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
		addVersionItem(versionItem, title, v.Identifier, item)
	}
	candidate[title] = versionMenu
}

func addVersionItem(item *systray.MenuItem, title string, version string, parentItem *systray.MenuItem) {
	installItem := item.AddSubMenuItem("Install && Use", "")
	uninstallItem := item.AddSubMenuItem("Uninstall", "")
	go func() {
		for {
			select {
			case <-installItem.ClickedCh:
				internal.UseCandidate(title, version, sdkmanInitScript)
				for _, v := range candidate[title] {
					if v.MenuItem != item {
						v.MenuItem.Uncheck()
					} else {
						item.Check()
					}
				}
			case <-uninstallItem.ClickedCh:
				internal.UninstallCandidate(title, version, sdkmanInitScript)
				item.SetTitle(version)
			}
		}
	}()
}

//func getIcon(s string) []byte {
//	// Read your icon file here
//	// For example, load from a file:
//	file, err := os.ReadFile(s)
//	if err != nil {
//		panic(err)
//	}
//	return file
//}
