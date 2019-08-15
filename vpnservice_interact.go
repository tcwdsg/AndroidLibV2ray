package libv2ray

import (
	"log"

	"golang.org/x/sys/unix"
)

/*VpnSupportReady VpnSupportReady*/
func (v *V2RayPoint) VpnSupportReady() {
	if !v.VpnSupportnodup {
		v.VpnSupportnodup = true
		v.VpnSupportSet.Setup(v.conf.vpnConfig.VPNSetupArg)
		v.setV2RayDialer()
		v.startVPNRequire()
	}
}
func (v *V2RayPoint) startVPNRequire() {
	go v.escortRun(v.conf.vpnConfig.Target, v.conf.vpnConfig.Args, false, v.VpnSupportSet.GetVPNFd())
}

func (v *V2RayPoint) askSupportSetInit() {
	v.VpnSupportSet.Prepare()
}

func (v *V2RayPoint) vpnSetup() {
	log.Println(v.conf.vpnConfig.VPNSetupArg)
	if v.conf.vpnConfig.VPNSetupArg != "" {
		v.askSupportSetInit()
	}
}
func (v *V2RayPoint) vpnShutdown() {

	if v.conf.vpnConfig.VPNSetupArg != "" {
		if v.VpnSupportnodup {
			unix.Close(v.VpnSupportSet.GetVPNFd())
		}
		v.VpnSupportSet.Shutdown()
	}
	v.VpnSupportnodup = false
}

func (v *V2RayPoint) setV2RayDialer() {
	//protectedDialer := &vpnProtectedDialer{vp: v}
	//@TODO 不清楚作用，新版已经没有
	//internet.SubstituteDialer(protectedDialer)
}
