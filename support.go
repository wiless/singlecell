package main

import (
	"fmt"

	"os"
	"path/filepath"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	cell "github.com/wiless/cellular"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
	"gonum.org/v1/gonum/stat"
)

func SwitchBack() {
	// pwd, _ := os.Getwd()
	// pwd, _ = filepath.EvalSymlinks(pwd)
	// cwd, _ := filepath.EvalSymlinks(currentdir)
	// rel, _ := filepath.Rel(pwd, cwd)
	// if pwd == cwd {
	// 	rel = "."
	// }
	// log.Printf("Switching Back", rel)
	os.Chdir(currentdir)
}

func SwitchInput() {
	pwd, _ := os.Getwd()
	currentdir = pwd
	rel, _ := filepath.Rel(currentdir, indir)
	log.Printf("Switching to INPUT DIR ./%s", rel)
	os.Chdir(indir)

}
func SwitchOutput() {
	pwd, _ := os.Getwd()
	currentdir = pwd
	rel, _ := filepath.Rel(currentdir, outdir)
	log.Printf("Switching to OUTPUT DIR ./%s", rel)
	os.Chdir(outdir)
}

func startMScript(fname string) *vlib.Matlab {
	SwitchOutput()
	m := vlib.NewMatlab("deployment")
	SwitchBack()
	m.Silent = true
	m.Json = false
	return m
}

// AttachAntennas attaches antennas to the nodes of type base-stations
func AttachAntennas(system deployment.DropSystem, bsids vlib.VectorI) {

	generateMScript := true
	if systemAntennas == nil {
		systemAntennas = make(map[int]*antenna.SettingAAS)
	}

	var m *vlib.Matlab

	if generateMScript {
		m = startMScript("deployment")
		m.Command("figure")
		cmd := `delta=pi/180;
		phaseangle=0:delta:2*pi-delta;`
		m.Command(cmd)
	}

	for _, i := range bsids {

		systemAntennas[i] = antenna.NewAAS()
		*systemAntennas[i] = defaultAAS
		systemAntennas[i].HTiltAngle = system.Nodes[i].Direction

		systemAntennas[i].CreateElements(system.Nodes[i].Location)
		hgain := vlib.NewVectorF(360)
		cnt := 0
		for d := 0; d < 360; d++ {
			hgain[cnt] = systemAntennas[i].ElementDirectionHGain(float64(d))
			cnt++
		}
		if generateMScript {
			m.Export("gain"+strconv.Itoa(i), hgain)
			cmd := fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
			m.Command(cmd)
		}
	}
	m.Close()
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
	rel, _ := filepath.Rel(defaultdir, defaultdir)
	log.Printf("WORK directory : %s", defaultdir)
	rel, _ = filepath.Rel(defaultdir, indir)
	log.Printf("INPUT directory :  ./%s", rel)
	rel, _ = filepath.Rel(defaultdir, outdir)
	log.Printf("OUTPUT directory :  ./%s", rel)

	// Read other parameters of the Application

}
func loadDefaults() {
	/// START OTHER THINGS
	defaultAAS.SetDefault()

	// defaultAAS.N = 1
	defaultAAS.FreqHz = CarriersGHz[0]
	// defaultAAS.BeamTilt = 0
	// defaultAAS.DisableBeamTit = false
	// defaultAAS.VTiltAngle = VTILT
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

// PrintCalibration prints the column of data used in calibration table of 3GPP
func PrintCalibration(metric map[int]cell.LinkMetric, rxids vlib.VectorI, fname string) {
	SwitchOutput()
	CFX := vlib.ToVectorF("0:100")

	CLsamples := vlib.NewVectorF(len(rxids))
	SINRsamples := vlib.NewVectorF(len(rxids))
	for i, rxid := range rxids {
		CLsamples[i] = metric[rxid].BestRSRP
		SINRsamples[i] = metric[rxid].BestSINR
	}
	CLsamples = CLsamples.Sorted()
	SINRsamples = SINRsamples.Sorted()
	vCouplingLoss := vlib.NewVectorF(CFX.Len())
	vSINR := vlib.NewVectorF(CFX.Len())
	mscript := vlib.NewMatlab("calibration")
	fid, _ := os.Create(fname)
	fmt.Fprintf(fid, "%% CDF\tCouplingGain\tSINR")
	for i, cfx := range CFX {
		vSINR[i] = stat.Quantile(cfx/100, stat.Empirical, SINRsamples, nil)
		vCouplingLoss[i] = stat.Quantile(cfx/100, stat.Empirical, CLsamples, nil)
		fmt.Fprintf(fid, "\n%f\t%f\t%f", cfx, vCouplingLoss[i], vSINR[i])
	}
	fid.Close()
	plotCDF(vCouplingLoss)
	mscript.Silent = true
	defer mscript.Close()
	mscript.Export("CFX", CFX)
	mscript.Export("CouplingLoss", vCouplingLoss)
	mscript.Export("SINR", vSINR)
	SwitchBack()

}

func plotCDF(v vlib.VectorF) {
	var Fx vlib.Vector2D
	Fx.Y, Fx.X = CDF(v)

	// Make a plot and set its title.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "CDF"

	// Create a histogram of our values drawn
	// from the standard normal.
	l, _, err := plotter.NewLinePoints(Fx)
	if err != nil {
		panic(err)
	}
	// Normalize the area under the histogram to
	// sum to one.
	p.Add(l)
	p.Add(plotter.NewGrid())

	// Save the plot to a PNG file.
	if err := p.Save(5*vg.Inch, 5*vg.Inch, "hist.svg"); err != nil {
		panic(err)
	}
	// Quantile
}

//Returns the CDF of the input vector v,
//Input may not be sorted, so it is sorted internally
func CDF(v vlib.VectorF) (Fx, x vlib.VectorF) {
	result := v.Sorted()
	rangestr := fmt.Sprintf("%f:%f", result[0], result[len(v)-1])
	couplingLoss := vlib.ToVectorF(rangestr)
	cdf := vlib.NewVectorF(couplingLoss.Len())
	// i := 0
	for i, q := range couplingLoss {
		val := stat.CDF(q, stat.Empirical, result, nil)
		cdf[i] = val
	}
	return cdf, couplingLoss
}

func printCDF(v vlib.VectorF) {
	result := v.Sorted()
	couplingLoss := vlib.ToVectorF("-160:-20")
	cdf := vlib.NewVectorF(couplingLoss.Len())
	// i := 0
	for i, q := range couplingLoss {
		val := stat.CDF(q, stat.Empirical, result, nil)
		cdf[i] = val
	}
	SwitchOutput()
	matlab2 := vlib.NewMatlab("calibration")
	matlab2.Silent = true
	SwitchBack()

	defer matlab2.Close()
	matlab2.Export("couplingGain", couplingLoss)
	matlab2.Export("cdf", cdf)
	matlab2.Command("open CalibrationResults.fig")

	CLGain := vlib.NewVectorF(101)
	CFX := vlib.ToVectorF("0:100")

	for i, cfx := range CFX {
		val := stat.Quantile(cfx/100, stat.Empirical, result, nil)
		CLGain[i] = val

	}
	matlab2.Export("CLGain", CLGain)
	matlab2.Export("cfx", vlib.ToVectorF("0:100"))

	matlab2.Command("hold all;")
	matlab2.Command("plot(couplingGain,cdf)")

	// Quantile
}

func DebugAntennaPattern() {
	SwitchOutput()
	mscript := vlib.NewMatlab("pattern")
	mscript.Silent = true
	defer mscript.Close()
	SwitchBack()

	azimuth := vlib.ToVectorF("0:360")
	elevation := vlib.ToVectorF("0:180")

	var HGain vlib.VectorF
	_ = elevation
	for i, theta := range azimuth {
		az, _, hgain := antenna.BSPatternDb(theta, 90.0+C.AntennaVTilt)
		HGain.AppendAtEnd(hgain)
		azimuth[i] = az
		// elevation[i] = el
	}

	mscript.Export("azimuth", azimuth)
	mscript.Export("azimuthGain", HGain)
	mscript.Command("figure;")
	mscript.Command("polar(deg2rad(azimuth),db2pow(azimuthGain),'r.');")
	mscript.Command("figure;")
	mscript.Command("plot(azimuth,(azimuthGain),'r.');")

}

func testCircular() {
	pts := deployment.AnnularRingEqPoints(deployment.ORIGIN, 100, 30)
	fmt.Printf("pos=%v", pts)
}
