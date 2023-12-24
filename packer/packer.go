//go:build !windows

package packer

func PlatformVars() (string, string) {
	const (
		url = "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_linux_amd64.zip"
		app = "packer"
	)
	return url, app
}
