// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Command releaselet does buildlet-side release construction tasks.
// It is intended to be executed on the buildlet preparing a release.
package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func main() {
	if err := blog(); err != nil {
		log.Fatal(err)
	}
	if err := tour(); err != nil {
		log.Fatal(err)
	}
	var err error
	switch runtime.GOOS {
	case "windows":
		err = windowsMSI()
	case "darwin":
		err = darwinPKG()
	}
	if err != nil {
		log.Fatal(err)
	}
}

const blogPath = "golang.org/x/blog"

var blogContent = []string{
	"content",
	"template",
}

func blog() error {
	// Copy blog content to $GOROOT/blog.
	blogSrc := filepath.Join("gopath/src", blogPath)
	contentDir := filepath.FromSlash("go/blog")
	return cpAllDir(contentDir, blogSrc, blogContent...)
}

const tourPath = "golang.org/x/tour"

var tourContent = []string{
	"content",
	"solutions",
	"static",
	"template",
}

var tourPackages = []string{
	"pic",
	"reader",
	"tree",
	"wc",
}

func tour() error {
	tourSrc := filepath.Join("gopath/src", tourPath)
	contentDir := filepath.FromSlash("go/misc/tour")

	// Copy all the tour content to $GOROOT/misc/tour.
	if err := cpAllDir(contentDir, tourSrc, tourContent...); err != nil {
		return err
	}

	// Copy the tour source code so it's accessible with $GOPATH pointing to $GOROOT/misc/tour.
	tourPKGDir := filepath.Join(contentDir, "src", tourPath)
	if err := cpAllDir(tourPKGDir, tourSrc, tourPackages...); err != nil {
		return err
	}

	// Copy gotour binary to tool directory as "tour"; invoked as "go tool tour".
	return cp(
		filepath.FromSlash("go/pkg/tool/"+runtime.GOOS+"_"+runtime.GOARCH+"/tour"+ext()),
		filepath.FromSlash("gopath/bin/gotour"+ext()),
	)
}

func environ() (cwd, version string, err error) {
	cwd, err = os.Getwd()
	if err != nil {
		return
	}
	var versionBytes []byte
	versionBytes, err = ioutil.ReadFile("go/VERSION")
	if err != nil {
		return
	}
	version = string(bytes.TrimSpace(versionBytes))
	return
}

func darwinPKG() error {
	cwd, version, err := environ()
	if err != nil {
		return err
	}

	// Write out darwin data that is used by the packaging process.
	defer os.RemoveAll("darwin")
	if err := writeDataFiles(darwinData, "darwin"); err != nil {
		return err
	}

	// Create a work directory and place inside the files as they should
	// be on the destination file system.
	work := filepath.Join(cwd, "darwinpkg")
	if err := os.MkdirAll(work, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(work)

	// Write out /etc/paths.d/go.
	const pathsBody = "/usr/local/go/bin"
	pathsDir := filepath.Join(work, "etc/paths.d")
	pathsFile := filepath.Join(pathsDir, "go")
	if err := os.MkdirAll(pathsDir, 0755); err != nil {
		return err
	}
	if err = ioutil.WriteFile(pathsFile, []byte(pathsBody), 0644); err != nil {
		return err
	}

	// Copy Go installation to /usr/local/go.
	goDir := filepath.Join(work, "usr/local/go")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		return err
	}
	if err := cpDir(goDir, "go"); err != nil {
		return err
	}

	// Build the package file.
	dest := "package"
	if err := os.Mkdir(dest, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(dest)

	if err := run("pkgbuild",
		"--identifier", "com.googlecode.go",
		"--version", version,
		"--scripts", "darwin/scripts",
		"--root", work,
		filepath.Join(dest, "com.googlecode.go.pkg"),
	); err != nil {
		return err
	}

	const pkg = "pkg" // known to cmd/release
	if err := os.Mkdir(pkg, 0755); err != nil {
		return err
	}
	return run("productbuild",
		"--distribution", "darwin/Distribution",
		"--resources", "darwin/Resources",
		"--package-path", dest,
		filepath.Join(cwd, pkg, "go.pkg"), // file name irrelevant
	)
}

func windowsMSI() error {
	cwd, version, err := environ()
	if err != nil {
		return err
	}

	// Install Wix tools.
	wix := filepath.Join(cwd, "wix")
	defer os.RemoveAll(wix)
	if err := installWix(wix); err != nil {
		return err
	}

	// Write out windows data that is used by the packaging process.
	win := filepath.Join(cwd, "windows")
	defer os.RemoveAll(win)
	if err := writeDataFiles(windowsData, win); err != nil {
		return err
	}

	// Gather files.
	goDir := filepath.Join(cwd, "go")
	appfiles := filepath.Join(win, "AppFiles.wxs")
	if err := runDir(win, filepath.Join(wix, "heat"),
		"dir", goDir,
		"-nologo",
		"-gg", "-g1", "-srd", "-sfrag",
		"-cg", "AppFiles",
		"-template", "fragment",
		"-dr", "INSTALLDIR",
		"-var", "var.SourceDir",
		"-out", appfiles,
	); err != nil {
		return err
	}

	// Build package.
	if err := runDir(win, filepath.Join(wix, "candle"),
		"-nologo",
		"-dGoVersion="+version,
		"-dWixGoVersion="+wixVersion(version),
		"-dArch="+runtime.GOARCH,
		"-dSourceDir="+goDir,
		filepath.Join(win, "installer.wxs"),
		appfiles,
	); err != nil {
		return err
	}

	msi := filepath.Join(cwd, "msi") // known to cmd/release
	if err := os.Mkdir(msi, 0755); err != nil {
		return err
	}
	return runDir(win, filepath.Join(wix, "light"),
		"-nologo",
		"-dcl:high",
		"-ext", "WixUIExtension",
		"-ext", "WixUtilExtension",
		"AppFiles.wixobj",
		"installer.wixobj",
		"-o", filepath.Join(msi, "go.msi"), // file name irrelevant
	)
}

const wixBinaries = "https://storage.googleapis.com/go-builder-data/wix35-binaries.zip"

// installWix fetches and installs the wix toolkit to the specified path.
func installWix(path string) error {
	// Fetch wix binary zip file.
	body, err := httpGet(wixBinaries)
	if err != nil {
		return err
	}

	// Unzip to path.
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		name := filepath.FromSlash(f.Name)
		err := os.MkdirAll(filepath.Join(path, filepath.Dir(name)), 0755)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(path, name), b, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func httpGet(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, errors.New(r.Status)
	}
	return body, nil
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func runDir(dir, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func cp(dst, src string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	fi, err := sf.Stat()
	if err != nil {
		return err
	}
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()
	// Windows doesn't implement Fchmod.
	if runtime.GOOS != "windows" {
		if err := df.Chmod(fi.Mode()); err != nil {
			return err
		}
	}
	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}
	if err := df.Close(); err != nil {
		return err
	}
	// Ensure the destination has the same mtime as the source.
	return os.Chtimes(dst, fi.ModTime(), fi.ModTime())
}

func cpDir(dst, src string) error {
	walk := func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, srcPath[len(src):])
		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return cp(dstPath, srcPath)
	}
	return filepath.Walk(src, walk)
}

func cpAllDir(dst, basePath string, dirs ...string) error {
	for _, dir := range dirs {
		if err := cpDir(filepath.Join(dst, dir), filepath.Join(basePath, dir)); err != nil {
			return err
		}
	}
	return nil
}

func ext() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

var versionRe = regexp.MustCompile(`^go(\d+(\.\d+)*)`)

// The Microsoft installer requires version format major.minor.build
// (http://msdn.microsoft.com/en-us/library/aa370859%28v=vs.85%29.aspx).
// Where the major and minor field has a maximum value of 255 and build 65535.
// The offical Go version format is goMAJOR.MINOR.PATCH at $GOROOT/VERSION.
// It's based on the Mercurial tag. Remove prefix and suffix to make the
// installer happy.
func wixVersion(v string) string {
	m := versionRe.FindStringSubmatch(v)
	if m == nil {
		return "0.0.0"
	}
	return m[1]
}

const storageBase = "https://storage.googleapis.com/go-builder-data/release/"

// writeDataFiles writes the files in the provided map to the provided base
// directory. If the map value is a URL it fetches the data at that URL and
// uses it as the file contents.
func writeDataFiles(data map[string]string, base string) error {
	for name, body := range data {
		dst := filepath.Join(base, name)
		err := os.MkdirAll(filepath.Dir(dst), 0755)
		if err != nil {
			return err
		}
		b := []byte(body)
		if strings.HasPrefix(body, storageBase) {
			b, err = httpGet(body)
			if err != nil {
				return err
			}
		}
		// (We really mean 0755 on the next line; some of these files
		// are executable, and there's no harm in making them all so.)
		if err := ioutil.WriteFile(dst, b, 0755); err != nil {
			return err
		}
	}
	return nil
}

var darwinData = map[string]string{

	"scripts/postinstall": `#!/bin/bash
GOROOT=/usr/local/go
echo "Fixing permissions"
cd $GOROOT
find . -exec chmod ugo+r \{\} \;
find bin -exec chmod ugo+rx \{\} \;
find . -type d -exec chmod ugo+rx \{\} \;
chmod o-w .
`,

	"scripts/preinstall": `#!/bin/bash
GOROOT=/usr/local/go
echo "Removing previous installation"
if [ -d $GOROOT ]; then
	rm -r $GOROOT
fi
`,

	"Distribution": `<?xml version="1.0" encoding="utf-8" standalone="no"?>
<installer-script minSpecVersion="1.000000">
    <title>Go</title>
    <background mime-type="image/png" file="bg.png"/>
    <options customize="never" allow-external-scripts="no"/>
    <domains enable_localSystem="true" />
    <installation-check script="installCheck();"/>
    <script>
function installCheck() {
    if(!(system.compareVersions(system.version.ProductVersion, '10.6.0') >= 0)) {
        my.result.title = 'Unable to install';
        my.result.message = 'Go requires Mac OS X 10.6 or later.';
        my.result.type = 'Fatal';
        return false;
    }
    if(system.files.fileExistsAtPath('/usr/local/go/bin/go')) {
	    my.result.title = 'Previous Installation Detected';
	    my.result.message = 'A previous installation of Go exists at /usr/local/go. This installer will remove the previous installation prior to installing. Please back up any data before proceeding.';
	    my.result.type = 'Warning';
	    return false;
	}
    return true;    
}
    </script>
    <choices-outline>
        <line choice="com.googlecode.go.choice"/>
    </choices-outline>
    <choice id="com.googlecode.go.choice" title="Go">
        <pkg-ref id="com.googlecode.go.pkg"/>
    </choice>
    <pkg-ref id="com.googlecode.go.pkg" auth="Root">com.googlecode.go.pkg</pkg-ref>
</installer-script>
`,

	"Resources/bg.png": storageBase + "darwin/bg.png",
}

var windowsData = map[string]string{

	"installer.wxs": `<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
<!--
# Copyright 2010 The Go Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
-->

<?if $(var.Arch) = 386 ?>
  <?define ProdId = {FF5B30B2-08C2-11E1-85A2-6ACA4824019B} ?>
  <?define UpgradeCode = {1C3114EA-08C3-11E1-9095-7FCA4824019B} ?>
  <?define SysFolder=SystemFolder ?>
<?else?>
  <?define ProdId = {716c3eaa-9302-48d2-8e5e-5cfec5da2fab} ?>
  <?define UpgradeCode = {22ea7650-4ac6-4001-bf29-f4b8775db1c0} ?>
  <?define SysFolder=System64Folder ?>
<?endif?>

<Product
    Id="*"
    Name="Go Programming Language $(var.Arch) $(var.GoVersion)"
    Language="1033"
    Version="$(var.WixGoVersion)"
    Manufacturer="https://golang.org"
    UpgradeCode="$(var.UpgradeCode)" >

<Package
    Id='*' 
    Keywords='Installer'
    Description="The Go Programming Language Installer"
    Comments="The Go programming language is an open source project to make programmers more productive."
    InstallerVersion="300"
    Compressed="yes"
    InstallScope="perMachine"
    Languages="1033" />
    <!--    Platform="x86 or x64" -->

<Property Id="ARPCOMMENTS" Value="The Go programming language is a fast, statically typed, compiled language that feels like a dynamically typed, interpreted language." />
<Property Id="ARPCONTACT" Value="golang-nuts@googlegroups.com" />
<Property Id="ARPHELPLINK" Value="https://golang.org/help/" />
<Property Id="ARPREADME" Value="https://golang.org" />
<Property Id="ARPURLINFOABOUT" Value="https://golang.org" />
<Property Id="LicenseAccepted">1</Property>
<Icon Id="gopher.ico" SourceFile="images\gopher.ico"/>
<Property Id="ARPPRODUCTICON" Value="gopher.ico" />
<Media Id='1' Cabinet="go.cab" EmbedCab="yes" CompressionLevel="high" />
<Condition Message="Windows 2000 or greater required."> VersionNT >= 500</Condition>
<MajorUpgrade AllowDowngrades="yes" />
<SetDirectory Id="INSTALLDIRROOT" Value="[%SYSTEMDRIVE]"/>

<CustomAction
    Id="SetApplicationRootDirectory"
    Property="ARPINSTALLLOCATION"
    Value="[INSTALLDIR]" />

<!-- Define the directory structure and environment variables -->
<Directory Id="TARGETDIR" Name="SourceDir">
  <Directory Id="INSTALLDIRROOT">
    <Directory Id="INSTALLDIR" Name="Go"/>
  </Directory>
  <Directory Id="ProgramMenuFolder">
    <Directory Id="GoProgramShortcutsDir" Name="Go Programming Language"/>
  </Directory>
  <Directory Id="EnvironmentEntries">
    <Directory Id="GoEnvironmentEntries" Name="Go Programming Language"/>
  </Directory>
</Directory>

<!-- Programs Menu Shortcuts -->
<DirectoryRef Id="GoProgramShortcutsDir">
  <Component Id="Component_GoProgramShortCuts" Guid="{f5fbfb5e-6c5c-423b-9298-21b0e3c98f4b}">
    <Shortcut
        Id="GoDocServerStartMenuShortcut"
        Name="GoDocServer"
        Description="Starts the Go documentation server (http://localhost:6060)"
        Show="minimized"
        Arguments='/c start "Godoc Server http://localhost:6060" "[INSTALLDIR]bin\godoc.exe" -http=localhost:6060 -goroot="[INSTALLDIR]." &amp;&amp; start http://localhost:6060'
        Icon="gopher.ico"
        Target="[%ComSpec]" />
    <Shortcut
        Id="UninstallShortcut"
        Name="Uninstall Go"
        Description="Uninstalls Go and all of its components"
        Target="[$(var.SysFolder)]msiexec.exe"
        Arguments="/x [ProductCode]" />
    <RemoveFolder
        Id="GoProgramShortcutsDir"
        On="uninstall" />
    <RegistryValue
        Root="HKCU"
        Key="Software\GoProgrammingLanguage"
        Name="ShortCuts"
        Type="integer" 
        Value="1"
        KeyPath="yes" /> 
  </Component>
</DirectoryRef>

<!-- Registry & Environment Settings -->
<DirectoryRef Id="GoEnvironmentEntries">
  <Component Id="Component_GoEnvironment" Guid="{3ec7a4d5-eb08-4de7-9312-2df392c45993}">
    <RegistryKey 
        Root="HKCU"
        Key="Software\GoProgrammingLanguage"
        Action="create" >
            <RegistryValue
                Name="installed"
                Type="integer"
                Value="1"
                KeyPath="yes" />
            <RegistryValue
                Name="installLocation"
                Type="string"
                Value="[INSTALLDIR]" />
    </RegistryKey>
    <Environment
        Id="GoPathEntry"
        Action="set"
        Part="last"
        Name="PATH"
        Permanent="no"
        System="yes"
        Value="[INSTALLDIR]bin" />
    <Environment
        Id="GoRoot"
        Action="set"
        Part="all"
        Name="GOROOT"
        Permanent="no"
        System="yes"
        Value="[INSTALLDIR]" />
    <RemoveFolder
        Id="GoEnvironmentEntries"
        On="uninstall" />
  </Component>
</DirectoryRef>

<!-- Install the files -->
<Feature
    Id="GoTools"
    Title="Go"
    Level="1">
      <ComponentRef Id="Component_GoEnvironment" />
      <ComponentGroupRef Id="AppFiles" />
      <ComponentRef Id="Component_GoProgramShortCuts" />
</Feature>

<!-- Update the environment -->
<InstallExecuteSequence>
    <Custom Action="SetApplicationRootDirectory" Before="InstallFinalize" />
</InstallExecuteSequence>

<!-- Include the user interface -->
<WixVariable Id="WixUILicenseRtf" Value="LICENSE.rtf" />
<WixVariable Id="WixUIBannerBmp" Value="images\Banner.jpg" />
<WixVariable Id="WixUIDialogBmp" Value="images\Dialog.jpg" />
<Property Id="WIXUI_INSTALLDIR" Value="INSTALLDIR" />
<UIRef Id="WixUI_InstallDir" />

</Product>
</Wix>
`,

	"LICENSE.rtf":           storageBase + "windows/LICENSE.rtf",
	"images/Banner.jpg":     storageBase + "windows/Banner.jpg",
	"images/Dialog.jpg":     storageBase + "windows/Dialog.jpg",
	"images/DialogLeft.jpg": storageBase + "windows/DialogLeft.jpg",
	"images/gopher.ico":     storageBase + "windows/gopher.ico",
}
