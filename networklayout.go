package main

import (
	"fmt"
	"math"
	"math/rand"

	log "github.com/Sirupsen/logrus"

	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

/// Calculate Pathloss
func DeployLayer1(system *deployment.DropSystem) {
	setting := system.GetSetting()

	if setting == nil {
		setting = deployment.NewDropSetting()

		GENERATE := true
		if GENERATE {

			BSHEIGHT := C.BSHeight
			UEHeight := C.UEHeight
			BSMode := deployment.TransmitOnly
			/// NodeType should come from API calls
			newnodetype := deployment.NodeType{Name: "BS0", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: C.NCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS1", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: C.NCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS2", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: C.NCells}
			newnodetype.Mode = BSMode
			setting.AddNodeType(newnodetype)

			UEMode := deployment.ReceiveOnly

			newnodetype = deployment.NodeType{Name: "UE", Hmin: UEHeight, Hmax: UEHeight, Count: C.NumUEperCell * C.ActiveUECells}
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

	// clocations := deployment.HexGrid(C.NCells, vlib.Origin3D, CellRadius, 30)
	clocations, vcells := deployment.HexWrapGrid(C.NCells, vlib.Origin3D, CellRadius, 30, 19)
	_ = vcells
	log.Infof("BS=(%d) %v..%v", C.NCells, clocations[0], clocations[C.NCells-1])
	log.Infof("vCells=%v", vcells)

	system.SetAllNodeLocation("BS0", vlib.Location3DtoVecC(clocations[0:C.NCells]))
	system.SetAllNodeLocation("BS1", vlib.Location3DtoVecC(clocations[0:C.NCells]))
	system.SetAllNodeLocation("BS2", vlib.Location3DtoVecC(clocations[0:C.NCells]))

	// Activate Transmission online in C.BScells
	sec0ID := system.GetNodeIDs("BS0")
	sec1ID := system.GetNodeIDs("BS1")
	sec2ID := system.GetNodeIDs("BS2")
	var actbsid vlib.VectorI
	actbsid = C.BScells
	var s0, s1, s2 int
	var n deployment.Node
	fmt.Println("Active Cell IDs ", C.BScells, actbsid)
	fmt.Println("Sector 0 Cell IDs ", sec0ID)
	for indx, cid := range sec0ID {
		if !actbsid.Contains(cid) {

			s0 = sec0ID[indx]
			n = system.Nodes[s0]
			n.Active = false
			system.Nodes[s0] = n

			s1 = sec1ID[indx]
			n = system.Nodes[s1]
			n.Active = false
			system.Nodes[s1] = n

			s2 = sec2ID[indx]
			n = system.Nodes[s2]
			n.Active = false
			system.Nodes[s2] = n

		}

	}

	system.SetAllNodeProperty("BS0", "VTilt", C.AntennaVTilt)
	system.SetAllNodeProperty("BS1", "VTilt", C.AntennaVTilt)
	system.SetAllNodeProperty("BS2", "VTilt", C.AntennaVTilt)

	// muelocations := LoadUELocations(system)
	muelocations := GenerateUELocations(C.UEcells)
	system.SetAllNodeLocation("UE", muelocations)

	ueids := system.GetNodeIDs("UE")

	// Set 40% of the UEs with Indoor, InCar
	gid := -1

	for indx, u := range ueids {
		n := system.Nodes[u]
		if math.Mod(float64(indx), float64(C.NumUEperCell)) == 0 {
			// fmt.Println("Current UE %d of Cell %d", u, C.UEcells[gid])
			gid++
		}
		n.GeoCellID = C.UEcells[gid]
		if rand.Float64() <= C.INDOORRatio {
			n.Indoor = true
		} else {
			n.Indoor = false
			n.InCar = false

			if C.INCARRatio > 0 {
				outdoorInCar := (C.INCARRatio / (1.0 - C.INDOORRatio))
				// Set INCARRatio of OUTDOOR as Incar
				if rand.Float64() <= outdoorInCar {
					n.InCar = true
				}
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

func GenerateUELocations(bsids vlib.VectorI) vlib.VectorC {
	deployment.MinDistance = 10 /// Atleast 10m away from center
	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(C.NCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	fmt.Println("Generating ", bsids)
	fmt.Println("Generating ", C.ActiveUECells)
	cnt := 0
	for indx, bsloc := range hexCenters {

		if bsids.Contains(indx) {
			cnt++
			log.Printf("Dropping Uniform %d UEs for cell %d", C.NumUEperCell, indx)
			ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, C.NumUEperCell, 30)
			// ulocation := deployment.CircularPoints(bsloc.Cmplx(), CellRadius, NMobileUEs)
			// ulocation := deployment.AnnularRingEqPoints(bsloc.Cmplx(), CellRadius/3, NMobileUEs)
			// ulocation := deployment.AnnularRingPoints(bsloc.Cmplx(), CellRadius/4, CellRadius*.75, NMobileUEs)
			// for i, v := range ulocation {
			// 	ulocation[i] = v + bsloc.Cmplx()
			// }
			uelocations = append(uelocations, ulocation...)
		}

	}
	if len(uelocations) != cnt*C.NumUEperCell {
		log.Warnf("GenerateUELocations : May be some UECells not found in BS", C.UEcells)
	}
	return uelocations
}

func LoadUELocations(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(activeUECells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)

	for indx, bsloc := range hexCenters {

		log.Printf("Dropping Uniform %d UEs for cell %d", NMobileUEs, indx)

		ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, NMobileUEs, 30)
		// ulocation := deployment.CircularPoints(bsloc.Cmplx(), CellRadius, NMobileUEs)
		// ulocation := deployment.AnnularRingEqPoints(bsloc.Cmplx(), CellRadius/3, NMobileUEs)
		// ulocation := deployment.AnnularRingPoints(bsloc.Cmplx(), CellRadius/4, CellRadius*2/3, NMobileUEs)
		// for i, v := range ulocation {
		// 	ulocation[i] = v + bsloc.Cmplx()
		// }
		uelocations = append(uelocations, ulocation...)
	}
	return uelocations

}
