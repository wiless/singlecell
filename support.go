package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

func SwitchBack() {
	pwd, _ := os.Getwd()
	log.Printf("Switching to DEFAULT %s to %s ", pwd, currentdir)
	os.Chdir(currentdir)
}

func SwitchInput() {
	pwd, _ := os.Getwd()
	currentdir = pwd
	log.Printf("Switching to INPUT %s to %s ", pwd, indir)
	os.Chdir(indir)

}
func SwitchOutput() {
	pwd, _ := os.Getwd()
	currentdir = pwd
	log.Printf("Switching to OUTPUT %s to %s ", pwd, outdir)
	os.Chdir(outdir)
}

// func LoadUELocationsV(system *deployment.DropSystem) vlib.VectorC {

// 	var uelocations vlib.VectorC
// 	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
// 	for indx, bsloc := range hexCenters {
// 		// log.Printf("Deployed for cell %d at %v", indx, bsloc.Cmplx())
// 		_ = indx
// 		// 3-Villages in the HEXAGONAL CELL
// 		//villageCentre := deployment.HexRandU(bsloc, CellRadius, NVillages, 30)

// 		// Practical
// 		//	villageCentres := deployment.AnnularRingPoints(bsloc.Cmplx(), 1500, 3000, NVillages)
// 		villageCentres := deployment.AnnularRingEqPoints(bsloc.Cmplx(), VillageDistance, NVillages) /// On
// 		offset := vlib.RandUFVec(NVillages).ShiftAndScale(0, 500.0)                                 // add U(0,1500)  scale by 1 to 2.0
// 		rotate := vlib.RandUFVec(NVillages).ScaleAndShift(math.Pi/10, -math.Pi/20)                  // +- 10 degrees
// 		_ = rotate
// 		_ = offset
// 		for v, vc := range villageCentres {
// 			// Add Random offset U(0,1500) Radially
// 			c := vc + cmplx.Rect(offset[v], cmplx.Phase(vc)) // +rotate[v]

// 			// log.Printf("Adding Village %d of GP %d , VC  %v , Radial Offset %v , %v, RESULT %v", v, indx, vc, offset[v], (cmplx.Phase(vc)), cmplx.Abs(c-vc))
// 			log.Printf("Adding Village %d of GP %d  : %d users", v, indx, NUEsPerVillage)
// 			villageUElocations := deployment.CircularPoints(c, VillageRadius, NUEsPerVillage)

// 			uelocations = append(uelocations, villageUElocations...)
// 		}

// 	}

// 	return uelocations
// }

// func LoadUELocationsGP(system *deployment.DropSystem) vlib.VectorC {

// 	var uelocations vlib.VectorC
// 	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
// 	for indx, bsloc := range hexCenters {
// 		log.Printf("Dropping GP %d UEs for cell %d", GPusers, indx)

// 		// AT GP
// 		uelocation := deployment.CircularPoints(bsloc.Cmplx(), GPradius, GPusers)
// 		uelocations = append(uelocations, uelocation...)

// 	}

// 	return uelocations

// }

func CreateAntennas(system deployment.DropSystem, bsids vlib.VectorI) {
	if systemAntennas == nil {
		systemAntennas = make(map[int]*antenna.SettingAAS)
	}

	// omni := antenna.NewAAS()
	// sector := antenna.NewAAS()

	// vlib.LoadStructure("omni.json", omni)
	// vlib.LoadStructure("sector.json", sector)

	for _, i := range bsids {

		systemAntennas[i] = antenna.NewAAS()
		// copy(systemAntennas[i], defaultAAS)
		// SwitchInput()
		// vlib.LoadStructure("sector.json", systemAntennas[i])
		// SwitchBack()
		*systemAntennas[i] = defaultAAS

		// systemAntennas[i].FreqHz = CarriersGHz[0] * 1.e9
		// systemAntennas[i].HBeamWidth = 65

		systemAntennas[i].HTiltAngle = system.Nodes[i].Direction

		// if nSectors == 1 {
		// 	systemAntennas[i].Omni = true
		// } else {
		// 	systemAntennas[i].Omni = false
		// }
		systemAntennas[i].CreateElements(system.Nodes[i].Location)
		// fmt.Printf("\nType=%s , BSid=%d : System Antenna : %v", system.Nodes[i].Type, i, systemAntennas[i].Centre)

		hgain := vlib.NewVectorF(360)
		// vgain := vlib.NewVectorF(360)

		cnt := 0
		cmd := `delta=pi/180;
		phaseangle=0:delta:2*pi-delta;`
		matlab.Command(cmd)
		for d := 0; d < 360; d++ {
			hgain[cnt] = systemAntennas[i].ElementDirectionHGain(float64(d))
			//		hgain[cnt] = systemAntennas[i].ElementEffectiveGain(thetaH, thetaV)
			cnt++
		}

		// SwitchOutput()
		matlab.Export("gain"+strconv.Itoa(i), hgain)
		// SwitchBack()
		// fmt.Printf("\nBS %d, Antenna : %#v", i, systemAntennas[i])

		cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
		matlab.Command(cmd)
	}
}

func ReadConfig() {

	defaultdir, _ = os.Getwd()
	currentdir = defaultdir
	if indir == "." {
		indir = defaultdir
	} else {
		finfo, err := os.Stat(indir)
		if err != nil {
			log.Println("Error Input Dir ", indir, err)
			os.Exit(-1)
		} else {
			if !finfo.IsDir() {
				log.Println("Error Input Dir is not a Directory ", indir)
				os.Exit(-1)
			}
		}

	}

	if outdir == "." {
		outdir = defaultdir
	} else {
		finfo, err := os.Stat(outdir)
		if err != nil {
			log.Print("Creating OUTPUT directory : ", outdir)
			err = os.Mkdir(outdir, os.ModeDir|os.ModePerm)
			if err != nil {
				log.Print("Error Creating Directory ", outdir, err)
				os.Exit(-1)
			}

		} else {
			if !finfo.IsDir() {
				log.Panicln("Error Output Dir is not a Directory ", outdir)
			}
		}

	}
	outdir, _ = filepath.Abs(outdir)
	indir, _ = filepath.Abs(indir)
	log.Printf("WORK directory : %s", defaultdir)
	log.Printf("INPUT directory :  %s", indir)
	log.Printf("OUTPUT directory :  %s", outdir)

	// Read other parameters of the Application

}
func loadDefaults() {
	/// START OTHER THINGS
	defaultAAS.SetDefault()

	// defaultAAS.N = 1
	defaultAAS.FreqHz = CarriersGHz[0]
	// defaultAAS.BeamTilt = 0
	// defaultAAS.DisableBeamTit = false
	defaultAAS.VTiltAngle = VTILT
	// defaultAAS.ESpacingVFactor = .5
	// defaultAAS.HTiltAngle = 0
	// defaultAAS.MfileName = "output.m"
	// defaultAAS.Omni = true
	// defaultAAS.GainDb = 10
	// defaultAAS.HoldOn = false
	// defaultAAS.AASArrayType = antenna.LinearPhaseArray
	// defaultAAS.CurveWidthInDegree = 30.0
	// defaultAAS.CurveRadius = 1.00

}
