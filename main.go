package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"

	"math"
	"math/rand"
	"os"
	"time"

	"github.com/namsral/flag"
	cell "github.com/wiless/cellular"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/channelmodel"
	"github.com/wiless/vlib"
)

var defaultAAS antenna.SettingAAS
var systemAntennas map[int]*antenna.SettingAAS

var singlecell deployment.DropSystem
var secangles = vlib.VectorF{0.0, 120.0, -120.0}

// var nUEPerCell = 1000
var nCells = 19

var trueCells int // The number of cells where the UEs are dropped ..

var ISD float64 = 1732
var CellRadius float64
var TxPowerDbm float64 = 46.0
var CarriersGHz = vlib.VectorF{0.7}
var RXTYPES = []string{"UE"}
var VTILT float64 = 15.0

var NMobileUEs = 10 // 100
var fnameSINRTable string
var fnameMetricName string
var outdir string
var indir string
var defaultdir string
var currentdir string
var rma CM.RMa

func init() {

	fs := flag.NewFlagSetWithEnvPrefix("CELL", "WILESS", flag.ExitOnError)
	fs.String(flag.DefaultConfigFlagname, "config.cfg", "path to config file")

	fs.StringVar(&outdir, "outdir", ".", "Directory where all the output files are generated..")
	fs.StringVar(&indir, "indir", ".", "Directory where all the input files are read..")
	help := fs.Bool("help", false, "prints this help")
	verbose := fs.Bool("v", true, "Print logs verbose mode")
	// fs.Parse(os.Args[0])
	fs.Parse(os.Args[1:])
	if *help {
		fs.PrintDefaults()
		os.Exit(0)
		return
	}

	ReadConfig()
	log.Println("Current indir & outdir ", indir, outdir)
	//	vlib.LoadStructure("omni.json", &defaultAAS)

	SwitchInput()
	vlib.LoadStructure("sector.json", &defaultAAS)
	C.AntennaVTilt = defaultAAS.VTiltAngle
	SwitchBack()

	// vlib.LoadStructure("omni.json", defaultAAS)
	ReadAppConfig()
	if C.TrueCells == -1 {
		C.TrueCells = nCells
	}
	trueCells = C.TrueCells

	defaultAAS.VTiltAngle = C.AntennaVTilt

	// vlib.SaveStructure(defaultAAS, "defaultAAS.json", true)

	fnameSINRTable = "table700MHz.dat"

	// fnameMetricName = "metric700MHz" + cast.ToString(TxPowerDbm) + cast.ToString(CellRadius) + ".json"
	fnameMetricName = "metric700MHz.json"
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
}

func main() {
	// log.SetLevel(log.InfoLevel)
	SwitchOutput()
	matlab = vlib.NewMatlab("deployment")
	SwitchBack()
	matlab.Silent = true
	matlab.Json = false
	log.Print("BS,UE NoiseFigure :", C.BSNoiseFigureDb, C.UENoiseFigureDb)

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	// seedvalue = 0
	_ = seedvalue
	// rand.Seed(seedvalue)

	// var plmodel pathloss.OkumuraHata
	// _ = plmodel
	// // var plmodel walfisch.WalfischIke
	// // var plmodel pathloss.SimplePLModel
	// // var plmodel pathloss.RMa

	rma.LoadIMT2020(CarriersGHz[0])

	if C.ForceAllLOS {
		rma.ForceAllLOS(true)

	}

	DeployLayer1(&singlecell)

	// singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	// singlecell.SetAllNodeProperty("UE", "AntennaType", 1)

	rxnodeTypes := singlecell.GetNodeTypesOfMode(deployment.ReceiveOnly)
	log.Println("Found these RX nodes ", rxnodeTypes)

	for _, uetype := range rxnodeTypes {
		singlecell.SetAllNodeProperty(uetype, "FreqGHz", CarriersGHz)
	}

	layerBS := []string{"BS0", "BS1", "BS2"}

	var bsids vlib.VectorI

	for indx, bs := range layerBS {
		singlecell.SetAllNodeProperty(bs, "FreqGHz", CarriersGHz)
		singlecell.SetAllNodeProperty(bs, "TxPowerDBm", TxPowerDbm)
		singlecell.SetAllNodeProperty(bs, "Direction", secangles[indx])
		newids := singlecell.GetNodeIDs(bs)
		bsids.AppendAtEnd(newids...)
		log.Printf("\n %s : %v", bs, newids)

	}

	CreateAntennas(singlecell, bsids)
	SwitchOutput()
	//	vlib.SaveStructure(systemAntennas, "antennaArray.json", true)
	// vlib.SaveStructure(singlecell.GetSetting(), "dep.json", true)
	// vlib.SaveStructure(singlecell.Nodes, "nodelist.json", true)

	rxtypes := RXTYPES

	/// DUMPING OUTPUT Databases

	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = C.BandwidthMHz
	wsystem.FrequencyGHz = CarriersGHz[0]

	rxids := singlecell.GetNodeIDs(rxtypes...)

	log.Println("Evaluating Link Gains for RXid range : ", rxids[0], rxids[len(rxids)-1], len(rxids))
	RxMetrics400 := make(map[int]cell.LinkMetric)
	baseCells := vlib.VectorI{0, 1, 2}
	// baseCells := vlib.NewSegmentI(0, nCells*3) // 3 sectors per cell
	baseCells = baseCells.Scale(nCells)

	// baseCells = baseCells.Scale(trueCells)
	wsystem.OtherLossFn = penetrationLossFn

	if C.ActiveCells == 1 {
		wsystem.ActiveCells = baseCells
	}
	// log.Println("ActivebaseCells ", baseCells)
	for _, rxid := range rxids {

		metric := wsystem.EvaluateLinkMetricV2(&singlecell, &rma, rxid, myfunc)
		// metric := wsystem.EvaluteLinkMetric(&singlecell, &plmodel, rxid, myfunc)

		RxMetrics400[rxid] = metric
	}

	SwitchOutput()
	{ // Dump UE locations

		fid, _ := os.Create("uelocations.dat")
		ueids := singlecell.GetNodeIDs(rxtypes...)
		log.Println("RXid range : ", ueids[0], ueids[len(ueids)-1], len(ueids))
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tIndoor\tInCar\tBSdist")
		for _, id := range ueids {
			node := singlecell.Nodes[id]
			var ii int
			var ic int
			if node.Indoor {
				ii = 1

			}
			if node.InCar {
				ic = 1
			}
			bestbs := RxMetrics400[id].BestRSRPNode
			src := singlecell.Nodes[bestbs].Location
			dist := src.DistanceFrom(node.Location)
			// _ = ii
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %d \t %d \t %f", id, node.Location.X, node.Location.Y, node.Location.Z, ii, ic, dist)

		}
		fid.Close()

	}

	{ // Dump bs nodelocations
		fid, _ := os.Create("bslocations.dat")

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tPower\tdirection")
		for _, id := range bsids {
			node := singlecell.Nodes[id]
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, node.Location.X, node.Location.Y, node.Location.Z, node.TxPowerDBm, node.Direction)

		}
		fid.Close()

	}
	{ // Dump antenna nodelocations
		fid, _ := os.Create("antennalocations.dat")
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tHDirection\tHWidth\tVTilt")
		for _, id := range bsids {
			ant := myfunc(id)
			// if id%7 == 0 {
			// 	node.TxPowerDBm = 0
			// } else {
			// 	node.TxPowerDBm = 44
			// }
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f \t %f", id, ant.Centre.X, ant.Centre.Y, ant.Centre.Z, ant.HTiltAngle, ant.HBeamWidth, ant.VTiltAngle)

		}
		fid.Close()
	}

	{ /// Evaluage Dominant Interference Profiles
		MAXINTER := 8
		var fnameDIP string

		fnameDIP = "DIPprofilesNORM"
		MatlabResult := EvaluateDIP(RxMetrics400, rxids, MAXINTER, true) // Evaluates the normalized Dominant Interference Profiles

		vlib.SaveStructure(MatlabResult, fnameDIP+".json", true)
		fnameDIP = "DIPprofiles"
		MatlabResult = EvaluateDIP(RxMetrics400, rxids, MAXINTER, false) // Evaluates the normalized Dominant Interference Profiles
		vlib.SaveStructure(MatlabResult, fnameDIP+".json", true)

	}

	vlib.DumpMap2CSV(fnameSINRTable, RxMetrics400)
	vlib.SaveStructure(RxMetrics400, fnameMetricName, true)
	SwitchBack()
	matlab.Close()

	SaveAppConfig()
	fmt.Println("\n ============================")

}

/// Calculate Pathloss

func DeployLayer1(system *deployment.DropSystem) {
	setting := system.GetSetting()

	if setting == nil {
		setting = deployment.NewDropSetting()

		GENERATE := true
		if GENERATE {

			BSHEIGHT := 35.0
			BSMode := deployment.TransmitOnly
			/// NodeType should come from API calls
			newnodetype := deployment.NodeType{Name: "BS0", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS1", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS2", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			UEMode := deployment.ReceiveOnly

			newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.5, Hmax: 1.5, Count: NMobileUEs * trueCells}
			newnodetype.Mode = UEMode
			setting.AddNodeType(newnodetype)

			// vlib.SaveStructure(setting, "depSettings.json", true)

		} else {
			SwitchInput()
			vlib.LoadStructure("depSettings.json", setting)
			SwitchBack()
			fmt.Printf("\n %#v", setting.NodeTypes)
		}
		system.SetSetting(setting)

	}

	system.Init()

	// Workaround else should come from API calls or Databases
	// bslocations := LoadBSLocations(system)
	// system.SetAllNodeLocation("BS", vlib.Location3DtoVecC(bslocations))

	// area := deployment.RectangularCoverage(600)
	// deployment.DropSetting.SetCoverage(area)

	// clocations := deployment.HexGrid(nCells, vlib.Origin3D, CellRadius, 30)
	clocations, vcells := deployment.HexWrapGrid(19, vlib.Origin3D, CellRadius, 30, 19)

	log.Infof("BS=[%v]", clocations, vcells)
	log.Infof("vCells=[%v]", vcells)

	system.SetAllNodeLocation("BS0", vlib.Location3DtoVecC(clocations[0:19]))
	system.SetAllNodeLocation("BS1", vlib.Location3DtoVecC(clocations[0:19]))
	system.SetAllNodeLocation("BS2", vlib.Location3DtoVecC(clocations[0:19]))

	muelocations := LoadUELocations(system)

	system.SetAllNodeLocation("UE", muelocations)

	// Set 40% of the UEs with Indoor
	ueids := system.GetNodeIDs("UE")

	for _, u := range ueids {
		n := system.Nodes[u]

		if rand.Float64() <= C.INDOORRatio {
			n.Indoor = true
		} else {
			n.Indoor = false
			n.InCar = false
			outdoorInCar := C.INCARRatio * (1.0 - C.INDOORRatio)
			// Set INCARRatio of OUTDOOR as Incar
			if rand.Float64() <= outdoorInCar {
				n.InCar = true
			}
		}
		system.Nodes[u] = n
	}
	// system.SetAllNodeProperty("UE", "Indoor", true)

	fmt.Println("Nof MUE ", len(muelocations))
	// ueids := system.GetNodeIDs("UE")
	// for _, n := range ueids {
	//
	// }

}

func LoadUELocations(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	for indx, bsloc := range hexCenters {
		log.Printf("Dropping Uniform %d UEs for cell %d", NMobileUEs, indx)

		ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, NMobileUEs, 30)
		// for i, v := range ulocation {
		// 	ulocation[i] = v + bsloc.Cmplx()
		// }
		uelocations = append(uelocations, ulocation...)
	}
	return uelocations

}

func myfunc(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	obj, ok := systemAntennas[nodeID]
	if !ok {
		log.Printf("No antenna created !! for %d ", nodeID)
		return defaultAAS
	} else {
		// fmt.Printf("\nNode %d , Omni= %v, Dirction=(H%v,V%v) and center is %v", nodeID, obj.Omni, obj.HTiltAngle, obj.VTiltAngle, obj.Centre)
		return *obj
	}
}

func EvaluateDIP(RxMetrics map[int]cell.LinkMetric, rxids vlib.VectorI, MAXINTER int, DONORM bool) []MatInfo {

	MatlabResult := make([]MatInfo, len(rxids))

	for indx, rxid := range rxids {
		metric := RxMetrics[rxid]
		var minfo MatInfo
		minfo.UserID = metric.RxNodeID
		minfo.SecID = int(math.Floor(float64(metric.BestRSRPNode) / float64(nCells)))
		minfo.BaseID = metric.BestRSRPNode

		if metric.TxNodeIDs.Size() < MAXINTER {
			MAXINTER = len(metric.TxNodeIDs) - 1
		}
		// log.Println("METRIC TxNodes ", metric.TxNodeIDs)
		minfo.IfStation = metric.TxNodeIDs[1 : MAXINTER+1] // the first entry is best
		var ifrssi vlib.VectorF
		ifrssi = metric.TxNodesRSRP[1:]

		if DONORM {
			minfo.RSSI = 0 // normalized
			ifrssi = ifrssi.Sub(metric.TxNodesRSRP[0])
		} else {
			minfo.RSSI = metric.TxNodesRSRP[0]

		}

		residual := ifrssi[MAXINTER:]
		residual = vlib.InvDbF(residual)
		ifrssi = ifrssi[0:MAXINTER]

		minfo.IfRSSI = ifrssi // the first entry is best
		minfo.ThermalNoise = metric.N0
		if DONORM {
			minfo.ThermalNoise -= metric.RSSI
		}
		minfo.SINR = metric.BestSINR
		if vlib.Sum(residual) > 0 {
			minfo.RestOfInterference = vlib.Db(vlib.Sum(residual))
		} else {
			minfo.RestOfInterference = -999999
		}

		src := singlecell.Nodes[minfo.BaseID].Location
		dest := singlecell.Nodes[minfo.UserID].Location
		dist := src.DistanceFrom(dest)
		minfo.Distance = dist
		//// test Inf
		_, err := json.Marshal(minfo)
		if err != nil {
			fmt.Printf("\nError %v \n UID %#v , ROI %v", err, minfo, vlib.Sum(residual))
		}
		MatlabResult[indx] = minfo
	}
	return MatlabResult

}

func penetrationLossFn(tx, rx deployment.Node) float64 {
	// var Out2IndoorLoss float64 = 13
	// var RxNoiseFigure float64 = 7
	// log.Print(tx.ID, tx.Indoor, rx.ID, rx.Indoor, NoiseFigureDb)
	var losses float64 = 0

	if rx.Indoor != tx.Indoor {
		// losses += C.Out2IndoorLossDb // Using model based value
		d2In := rand.Float64() * 10.0 // Uniform [0 to 10m]
		losses += rma.O2ILossDb(rma.FGHz(), d2In)

	}

	if rx.InCar {
		losses += CM.O2ICarLossDb() //C.INCARLossdB
	}

	if strings.Contains(rx.Type, "UE") {
		losses += C.UENoiseFigureDb
	}

	if strings.Contains(rx.Type, "BS") {
		losses += C.BSNoiseFigureDb
	}

	return losses

}
