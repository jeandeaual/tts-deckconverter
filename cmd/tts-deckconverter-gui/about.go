package main

import (
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"deckconverter/log"
)

const fyneLicense = `Copyright (C) 2018 Fyne.io developers (see AUTHORS)
All rights reserved.


Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of Fyne.io nor the names of its contributors may be
      used to endorse or promote products derived from this software without
      specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.`

func showLicenseWindow(app fyne.App) {
	licenseWindow := app.NewWindow("Licensing Information")

	okButton := widget.NewButton("OK", func() {
		licenseWindow.Close()
	})

	buttons := fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		okButton,
		layout.NewSpacer(),
	)
	licenseContainer := fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(700, 400)),
		widget.NewScrollContainer(widget.NewLabel(fyneLicense)),
	)
	content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		licenseContainer,
		layout.NewSpacer(),
		buttons,
	)

	licenseWindow.SetContent(content)
	licenseWindow.Show()
}

func showAboutWindow(app fyne.App) {
	aboutWindow := app.NewWindow("About")

	var aboutMsg strings.Builder

	aboutMsg.WriteString(appName)
	if len(version) > 0 {
		aboutMsg.WriteString(" version ")
		if isSHA1(version) {
			aboutMsg.WriteString(version[:7])
		} else {
			aboutMsg.WriteString(version)
		}
		aboutMsg.WriteString(".")
	}
	aboutMsg.WriteString("\n\nBuilt with Go version ")
	aboutMsg.WriteString(getGoVersion())
	fyneVersion, err := getModuleVersion("fyne.io/fyne")
	if err != nil {
		log.Error(err)
	} else {
		aboutMsg.WriteString(" and Fyne version ")
		aboutMsg.WriteString(fyneVersion)
	}

	if !buildTime.IsZero() {
		aboutMsg.WriteString(" on ")
		aboutMsg.WriteString(buildTime.Format(time.RFC3339))
		aboutMsg.WriteString(".")
	}

	aboutLabel := widget.NewLabel(aboutMsg.String())

	licenseButton := widget.NewButton("Licensing Information", func() {
		showLicenseWindow(app)
	})
	okButton := widget.NewButton("OK", func() {
		aboutWindow.Close()
	})

	buttons := fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		licenseButton,
		okButton,
		layout.NewSpacer(),
	)
	aboutContainer := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		aboutLabel,
	)
	content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		aboutContainer,
		layout.NewSpacer(),
		buttons,
	)

	aboutWindow.SetContent(content)
	aboutWindow.Show()
}
