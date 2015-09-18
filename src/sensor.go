package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/toorop/gopentsdb"
)

// Fonction principale, appelée au lancement du programme
func main() {
	var err error
	// Ces variables vont stocker differences d'usage des CPU/Core entre deux
	// mesures
	var diffUser, diffNice, diffSys, diffWait, diffIdle, diffSum uint64
	// Les variables suivant vont êtres utilisées pour stocker les stats entre
	// deux mesures
	var cStatsPrev map[string]CPUStats
	var netIOPrev *NetIO
	var disksIOPrev map[string]DiskIO

	// On va lire la configuration depuis l'environnement
	// Le nom d'hote de la machine
	hostname := os.Getenv("SENSOR_HOSTNAME")
	if hostname == "" {
		log.Fatalln("env var SENSOR_HOSTNAME does not exist")
	}

	// La période en secondes => le temps à attendre entre 2 mesures.
	period := int64(10)
	periodStr := os.Getenv("SENSOR_PERIOD")
	if periodStr != "" {
		period, err = strconv.ParseInt(periodStr, 10, 32)
	}

	// Tokens d'authentification
	username := os.Getenv("SENSOR_OPENSTDB_USERNAME")
	if username == "" {
		log.Fatalln("env var SENSOR_OPENSTDB_USERNAME does not exist")
	}
	password := os.Getenv("SENSOR_OPENSTDB_PASSWORD")
	if password == "" {
		log.Fatalln("env var SENSOR_OPENSTDB_PASSWORD does not exist")
	}

	// On initialise un nouveau client OpenSTDB
	OpenSTDBClient, err := openstdb.NewClient(openstdb.ClientConfig{
		Endpoint: "https://opentsdb.iot.runabove.io/api/put",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Boucle infinie.
	// A chaque tour on va faire une serie de mesures
	for {
		// Cette variable va contenir les points a pousser vers OpenSTDB
		points2Push := []openstdb.Point{}

		// On fait une pause  entre chaque mesure
		// pour les curieux le fait de la faire en début de boucle, nous permet de
		// ne pas ajouter cette instruction avant chaque "continue" en cas d'erreur
		time.Sleep(time.Duration(period) * time.Second)

		// On va chercher le load
		load, err := GetLoadAvg()
		if err != nil {
			log.Print("ERROR: ", err)
			continue
		}

		// On crée les trois points relatifs au load
		ptLoadCurrent := openstdb.NewPoint()
		ptLoadCurrent.Metric = "loadavg.current"
		ptLoadCurrent.Timestamp = load.timestamp
		ptLoadCurrent.Value = load.current
		ptLoadCurrent.Tags["host"] = hostname
		points2Push = append(points2Push, ptLoadCurrent)

		ptLoad5Min := openstdb.NewPoint()
		ptLoad5Min.Metric = "loadavg.5min"
		ptLoad5Min.Timestamp = load.timestamp
		ptLoad5Min.Value = load.avg5
		ptLoad5Min.Tags["host"] = hostname
		points2Push = append(points2Push, ptLoad5Min)

		ptLoad15Min := openstdb.NewPoint()
		ptLoad15Min.Metric = "loadavg.15min"
		ptLoad15Min.Timestamp = load.timestamp
		ptLoad15Min.Value = load.avg15
		ptLoad15Min.Tags["host"] = hostname
		points2Push = append(points2Push, ptLoad15Min)

		// Les stats cpu
		// cStatsPrev va stocker les stats cpu du test précédent. Ca va nous
		// permettre d'avoir des usage en % sur la periode donnée
		now := time.Now().Unix()
		cStats, err := GetCPUStats()
		if err != nil {
			log.Print("ERROR: ", err)
			continue
		}

		// On génere les point pour tous les cpu/cores
		for cpu, stats := range *cStats {
			// si on a des stats antérieures on fait un dif pour connaitre la conso de
			// chaque en pourcentage sur l'interval
			if &cStatsPrev != nil {
				if prevStats, ok := cStatsPrev[cpu]; ok {
					diffUser = stats.User - prevStats.User
					diffNice = stats.Nice - prevStats.Nice
					diffSys = stats.Sys - prevStats.Sys
					diffWait = stats.Wait - prevStats.Wait
					diffIdle = stats.Idle - prevStats.Idle
					diffSum = diffUser + diffNice + diffSys + diffWait + diffIdle

					// user
					point := openstdb.NewPoint()
					point.Metric = "cpu." + cpu + ".user.percent"
					point.Timestamp = now
					point.Value = float64(diffUser * 100 / diffSum)
					point.Tags["host"] = hostname
					points2Push = append(points2Push, point)

					// nice
					point = openstdb.NewPoint()
					point.Metric = "cpu." + cpu + ".nice.percent"
					point.Timestamp = now
					point.Value = float64(diffNice * 100 / diffSum)
					point.Tags["host"] = hostname
					points2Push = append(points2Push, point)

					// sys
					point = openstdb.NewPoint()
					point.Metric = "cpu." + cpu + ".sys.percent"
					point.Timestamp = now
					point.Value = float64(diffSys * 100 / diffSum)
					point.Tags["host"] = hostname
					points2Push = append(points2Push, point)

					// wait
					point = openstdb.NewPoint()
					point.Metric = "cpu." + cpu + ".wait.percent"
					point.Timestamp = now
					point.Value = float64(diffWait * 100 / diffSum)
					point.Tags["host"] = hostname
					points2Push = append(points2Push, point)

					// idle
					point = openstdb.NewPoint()
					point.Metric = "cpu." + cpu + ".idle.percent"
					point.Timestamp = now
					point.Value = float64(diffIdle * 100 / diffSum)
					point.Tags["host"] = hostname
					points2Push = append(points2Push, point)
				}
			}
		}
		cStatsPrev = *cStats

		// Mémoire
		now = time.Now().Unix()
		memStats, err := GetMemStats()
		if err == nil {
			// Free
			ptMemFree := openstdb.NewPoint()
			ptMemFree.Metric = "mem.free"
			ptMemFree.Timestamp = now
			ptMemFree.Value = float64(memStats["MemFree"])
			ptMemFree.Tags["host"] = hostname
			points2Push = append(points2Push, ptMemFree)

			// Memoire Utilisée
			ptMemUsed := openstdb.NewPoint()
			ptMemUsed.Metric = "mem.used"
			ptMemUsed.Timestamp = now
			ptMemUsed.Value = float64(memStats["MemTotal"] - memStats["MemFree"])
			ptMemUsed.Tags["host"] = hostname
			points2Push = append(points2Push, ptMemUsed)

			// Cached
			ptMemCached := openstdb.NewPoint()
			ptMemCached.Metric = "mem.cached"
			ptMemCached.Timestamp = now
			ptMemCached.Value = float64(memStats["Cached"])
			ptMemCached.Tags["host"] = hostname
			points2Push = append(points2Push, ptMemCached)

			// Buffers
			ptMemBuffers := openstdb.NewPoint()
			ptMemBuffers.Metric = "mem.buffers"
			ptMemBuffers.Timestamp = now
			ptMemBuffers.Value = float64(memStats["Buffers"])
			ptMemBuffers.Tags["host"] = hostname
			points2Push = append(points2Push, ptMemBuffers)

			// Swap free
			ptSwapFree := openstdb.NewPoint()
			ptSwapFree.Metric = "mem.swap.free"
			ptSwapFree.Timestamp = now
			ptSwapFree.Value = float64(memStats["SwapFree"])
			ptSwapFree.Tags["host"] = hostname
			points2Push = append(points2Push, ptSwapFree)

			// Swap used
			ptSwapused := openstdb.NewPoint()
			ptSwapused.Metric = "mem.swap.used"
			ptSwapused.Timestamp = now
			ptSwapused.Value = float64(memStats["SwapTotal"] - memStats["SwapFree"])
			ptSwapused.Tags["host"] = hostname
			points2Push = append(points2Push, ptSwapused)
		}

		// Le reseau
		io, err := GetNetIO()
		if err != nil {
			log.Println(err)
		} else if netIOPrev != nil {
			timeDelta := uint64(io.timestamp - netIOPrev.timestamp)

			// IN
			ptNetIn := openstdb.NewPoint()
			ptNetIn.Metric = "net.in"
			ptNetIn.Timestamp = io.timestamp
			ptNetIn.Value = float64((io.in - netIOPrev.in) / timeDelta)
			ptNetIn.Tags["host"] = hostname
			points2Push = append(points2Push, ptNetIn)

			// OUT
			ptNetOut := openstdb.NewPoint()
			ptNetOut.Metric = "net.out"
			ptNetOut.Timestamp = io.timestamp
			ptNetOut.Value = float64((io.out - netIOPrev.out) / timeDelta)
			ptNetOut.Tags["host"] = hostname
			points2Push = append(points2Push, ptNetOut)

		}
		netIOPrev = io

		// Les lectures / écritues disk
		disksIO, err := GetDisksIO()
		if err != nil {
			log.Println(err)
		} else if disksIOPrev != nil {
			for disk, stats := range disksIO {
				timeDelta := uint64(stats.Timestamp - disksIOPrev[disk].Timestamp)
				point := openstdb.NewPoint()
				point.Metric = "disk." + disk + ".reads"
				point.Timestamp = stats.Timestamp
				point.Value = float64((stats.Reads - disksIOPrev[disk].Reads) / timeDelta)
				point.Tags["host"] = hostname
				points2Push = append(points2Push, point)

				point = openstdb.NewPoint()
				point.Metric = "disk." + disk + ".writes"
				point.Timestamp = stats.Timestamp
				point.Value = float64((stats.Writes - disksIOPrev[disk].Writes) / timeDelta)
				point.Tags["host"] = hostname
				points2Push = append(points2Push, point)
			}
		}
		disksIOPrev = disksIO

		// On pousse les points vers OVH
		if err = OpenSTDBClient.Push(points2Push); err != nil {
			log.Println(err)
		}
	}
}
