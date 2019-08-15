package libv2ray

import (
	"os"
	//"v2ray.com/core/common/net"

	"v2ray.com/core"
	// The following are necessary as they register handlers in their init functions.
	//_ "v2ray.com/core/app/router/rules"
	"v2ray.com/core/common/log"
	//// The following are necessary as they register handlers in their init functions.
	//_ "v2ray.com/core/proxy/blackhole"
	//_ "v2ray.com/core/proxy/dokodemo"
	//_ "v2ray.com/core/proxy/freedom"
	//_ "v2ray.com/core/proxy/http"
	//_ "v2ray.com/core/proxy/shadowsocks"
	//_ "v2ray.com/core/proxy/socks"
	//_ "v2ray.com/core/proxy/vmess/inbound"
	//_ "v2ray.com/core/proxy/vmess/outbound"
	//
	//// The following are necessary as they register handlers in their init functions.
	//_ "v2ray.com/core/transport/internet/kcp"
	//_ "v2ray.com/core/transport/internet/tcp"
	//_ "v2ray.com/core/transport/internet/udp"
	//_ "v2ray.com/core/transport/internet/ws"
	//
	//// The following are necessary as they register handlers in their init functions.
	//_ "v2ray.com/core/transport/internet/authenticators/noop"
	//_ "v2ray.com/core/transport/internet/authenticators/srtp"
	//_ "v2ray.com/core/transport/internet/authenticators/utp"
)

/*V2RayPoint V2Ray Point Server
This is territory of Go, so no getter and setters!

Notice:
ConfigureFile can be either the path of config file or
"V2Ray_internal/ConfigureFileContent" in case you wish to

*/
type V2RayPoint struct {
	ConfigureFile        string
	ConfigureFileContent string
	Callbacks            V2RayCallbacks
	v2rServer            core.Server
	IsRunning            bool
	conf                 *libv2rayconf
	escortProcess        *[](*os.Process)
	unforgivnesschan     chan int
	VpnSupportSet        V2RayVPNServiceSupportsSet
	VpnSupportnodup      bool
	PackageName          string
	cfgtmpvarsi          cfgtmpvars
}

/*V2RayCallbacks a Callback set for V2Ray
 */
type V2RayCallbacks interface {
	OnEmitStatus(int, string) int
}

func (v *V2RayPoint) pointloop() {
	v.VpnSupportnodup = false

	if v.parseConf() != nil {
		return
	}

	err := v.checkIfRcExist()

	if err != nil {
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Error,
			Content:  "Failed to copy asset : " + err.Error(),
		})
		v.Callbacks.OnEmitStatus(-1, "Failed to copy asset ("+err.Error()+")")

	}

	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "v.renderAll() ",
	})
	v.renderAll()

	config, err := core.LoadConfig("json", "config.json", v.parseCfg())
	if err != nil {
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Error,
			Content:  "Failed to read config file (" + v.ConfigureFile + "): " + v.ConfigureFile + err.Error(),
		})

		v.Callbacks.OnEmitStatus(-1, "Failed to read config file ("+v.ConfigureFile+")")

		return
	}

	v2rServer, err := core.New(config)
	if err != nil {
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Error,
			Content:  "Failed to create Point server: " + err.Error(),
		})

		v.Callbacks.OnEmitStatus(-1, "Failed to create Point server ("+err.Error()+")")

		return
	}
	v.IsRunning = true
	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "vPoint.Start()",
	})
	v2rServer.Start()
	v.v2rServer = v2rServer

	log.Record(&log.GeneralMessage{
		Severity: log.Severity_Info,
		Content:  "vPoint.escortingUP()",
	})
	v.escortingUP()

	v.vpnSetup()

	if v.conf != nil {
		env := v.conf.additionalEnv
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Info,
			Content:  "Exec Upscript() "})
		err = v.runbash(v.conf.upscript, env)
		if err != nil {
			log.Record(&log.GeneralMessage{
				Severity: log.Severity_Error,
				Content:  "OnUp failed to exec: " + err.Error()})
		}
	}

	v.Callbacks.OnEmitStatus(0, "Running")
	v.parseCfgDone()
}

/*RunLoop Run V2Ray main loop
 */
func (v *V2RayPoint) RunLoop() {
	go v.pointloop()
}

func (v *V2RayPoint) stopLoopW() {
	v.IsRunning = false
	v.v2rServer.Close()

	if v.conf != nil {
		env := v.conf.additionalEnv
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Info,
			Content:  "Running downscript"})
		err := v.runbash(v.conf.downscript, env)

		if err != nil {
			log.Record(&log.GeneralMessage{
				Severity: log.Severity_Error,
				Content:  "OnDown failed to exec: " + err.Error()})
		}
		log.Record(&log.GeneralMessage{
			Severity: log.Severity_Info,
			Content:  "v.escortingDown() "})
		v.escortingDown()
	}

	v.Callbacks.OnEmitStatus(0, "Closed")

}

/*StopLoop Stop V2Ray main loop
 */
func (v *V2RayPoint) StopLoop() {
	v.vpnShutdown()
	go v.stopLoopW()
}

/*NewV2RayPoint new V2RayPoint*/
func NewV2RayPoint() *V2RayPoint {
	return &V2RayPoint{unforgivnesschan: make(chan int)}
}

/*NetworkInterrupted inform us to restart the v2ray,
closing dead connections.
*/
func (v *V2RayPoint) NetworkInterrupted() {
	v.v2rServer.Close()
	v.v2rServer.Start()
}
