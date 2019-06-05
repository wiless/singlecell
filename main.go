package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"os"

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
// var C.NCells = 61 // 19

var activeUECells int // The number of cells where the UEs are dropped ..

var ISD float64 = 1732
var CellRadius float64
var TxPowerDbm float64 = 46.0
var CarriersGHz = vlib.VectorF{3.5}
var RXTYPES = []string{"UE"}

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
	// log.Println("Current indir & outdir ", indir, outdir)
	//	vlib.LoadStructure("omni.json", &defaultAAS)

	SwitchInput()
	vlib.LoadStructure("sector.json", &defaultAAS)
	C.AntennaVTilt = defaultAAS.VTiltAngle
	SwitchBack()

	// vlib.LoadStructure("omni.json", defaultAAS)
	ReadAppConfig()
	if C.ActiveBSCells == -1 {
		C.ActiveBSCells = C.NCells
	}

	activeUECells = C.ActiveUECells

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

	if C.LogInfo == true {
		log.SetLevel(log.InfoLevel)
	}
	NMobileUEs = C.NumUEperCell
	log.Print("BS,UE NoiseFigure :", C.BSNoiseFigureDb, C.UENoiseFigureDb)

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	seedvalue = 0
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
	rma.ShadowLoss = C.ShadowLoss
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

	AttachAntennas(singlecell, bsids)
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
	// baseCells := vlib.NewSegmentI(0, C.NCells*3) // 3 sectors per cell
	baseCells = baseCells.Scale(C.NCells)

	// baseCells = baseCells.Scale(activeUECells)
	wsystem.OtherLossFn = penetrationLossFn

	if C.ActiveBSCells == 1 {
		wsystem.ActiveCells = baseCells
	}
	// log.Println("ActivebaseCells ", baseCells)
	for _, rxid := range rxids {
		metric := wsystem.EvaluateLinkMetricV2(&singlecell, &rma, rxid, antennaSelector)
		// metric := wsystem.EvaluteLinkMetric(&singlecell, &plmodel, rxid, myfunc)
		RxMetrics400[rxid] = metric
	}

	PrintCalibration(RxMetrics400, rxids, "calibration.dat")
	

	SwitchOutput()
	{ // Dump UE locations

		fid, _ := os.Create("uelocations.dat")
		ueids := singlecell.GetNodeIDs(rxtypes...)
		log.Println("RXid range : ", ueids[0], ueids[len(ueids)-1], len(ueids))
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tIndoor\tInCar\tBSdist\tGCellID")
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
			fmt.Fprintf(fid, "\n%d\t%f\t%f\t%f\t%d\t%d\t%f\t%d", id, node.Location.X, node.Location.Y, node.Location.Z, ii, ic, dist, node.GeoCellID)
			// couplingGain[indx] = RxMetrics400[id].BestRSRP
		}
		fid.Close()

	}

	{ // Dump bs nodelocations
		fid, _ := os.Create("bslocations.dat")

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tPower\tdirection\tActive")
		for _, id := range bsids {
			node := singlecell.Nodes[id]
			active := 0
			if node.Active {
				active = 1
			}
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f \t %d", id, node.Location.X, node.Location.Y, node.Location.Z, node.TxPowerDBm, node.Direction, active)

		}
		fid.Close()

	}
	{ // Dump antenna nodelocations
		fid, _ := os.Create("antennalocations.dat")
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tHDirection\tHWidth\tVTilt")
		for _, id := range bsids {
			ant := antennaSelector(id)
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
	SaveAppConfig()
	fmt.Println("\n ============================")

}

func antennaSelector(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	obj, ok := systemAntennas[nodeID]
	if !ok {
		log.Panicf("No antenna created !! for %d ", nodeID)
		return defaultAAS
	} else {
		// fmt.Printf("\nNode %d , Omni= %v, Direction=(H%v,V%v) and center is %v", nodeID, obj.Omni, obj.HTiltAngle, obj.VTiltAngle, obj.Centre)
		return *obj
	}
}

func penetrationLossFn(tx, rx deployment.Node) float64 {
	// var Out2IndoorLoss float64 = 13
	// var RxNoiseFigure float64 = 7
	// log.Print(tx.ID, tx.Indoor, rx.ID, rx.Indoor, NoiseFigureDb)
	var losses float64 = 0 // add a 8dB additionall loss

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
