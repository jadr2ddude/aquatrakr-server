package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type shower struct {
	len time.Duration
}

func (s shower) calcWater() float64 {
	return flowrate * s.len.Seconds()
}

func (s shower) genjson() (j showerjson) {
	j.TimeSeconds = s.len.Seconds()
	j.WaterUsed = s.calcWater()
	return
}

type showerjson struct {
	TimeSeconds float64
	WaterUsed   float64
}

var flowrate float64

var showers struct {
	sync.RWMutex
	s []shower
}

var motion struct {
	on    bool
	start time.Time
}

func main() {
	showers.s = []shower{
		shower{
			len: time.Second,
		},
		shower{
			len: time.Second * 2,
		},
	}
	flowrate = 1
	/*go func() {
		firmataAdaptor := firmata.NewAdaptor("/dev/ttyACM0")

		sensor := gpio.NewPIRMotionDriver(firmataAdaptor, "7")
		led := gpio.NewLedDriver(firmataAdaptor, "13")

		work := func() {
			sensor.On(gpio.MotionDetected, func(data interface{}) {
				//fmt.Println(gpio.MotionDetected)
				led.On()
				if !motion.on {
					motion.on = true
					motion.start = time.Now()
				}
			})
			sensor.On(gpio.MotionStopped, func(data interface{}) {
				//fmt.Println(gpio.MotionStopped)
				led.Off()
				if motion.on {
					motion.on = false
					showers.Lock()
					showers.s = append(showers.s, shower{len: time.Since(motion.start)})
					showers.Unlock()
				}
			})
		}

		robot := gobot.NewRobot("motionBot",
			[]gobot.Connection{firmataAdaptor},
			[]gobot.Device{sensor, led},
			work,
		)

		robot.Start()
	}()*/
	http.HandleFunc("/setFlow", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "wrong http method", http.StatusBadRequest)
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("parse error %s", err), http.StatusBadRequest)
		}
		flow := r.FormValue("flow")
		if flow == "" {
			http.Error(w, "missing flow value", http.StatusBadRequest)
		}
		fl, err := strconv.ParseFloat(flow, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("flow value parse error %s", err), http.StatusBadRequest)
		}
		flowrate = fl
	})
	http.HandleFunc("/addShower", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "wrong http method", http.StatusBadRequest)
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("parse error %s", err), http.StatusBadRequest)
		}
		t := r.FormValue("time")
		if t == "" {
			http.Error(w, "missing time value", http.StatusBadRequest)
		}
		ts, err := strconv.ParseFloat(t, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("time value parse error %s", err), http.StatusBadRequest)
		}
		showers.Lock()
		showers.s = append(showers.s, shower{time.Duration(ts * float64(time.Second))})
		showers.Unlock()
	})
	http.HandleFunc("/recent", func(w http.ResponseWriter, r *http.Request) {
		showers.RLock()
		defer showers.RUnlock()
		json.NewEncoder(w).Encode(showers.s[len(showers.s)-1].genjson())
	})
	http.HandleFunc("/getAll", func(w http.ResponseWriter, r *http.Request) {
		showers.RLock()
		defer showers.RUnlock()
		sh := make([]showerjson, len(showers.s))
		for i, v := range showers.s {
			sh[i] = v.genjson()
		}
		json.NewEncoder(w).Encode(sh)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		showers.RLock()
		defer showers.RUnlock()
		w.Header().Set("Expires", time.Unix(0, 0).Format(time.RFC1123))
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Accel-Expires", "0")
		fmt.Fprintln(w, `
	            <head>
	                <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.98.2/css/materialize.min.css">
	                <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.98.2/js/materialize.min.js"></script>
	            </head>
	            <body>
	                <nav class="blue">
	                    <div class="nav-wrapper">
	                        <a href="#" class="brand-logo">AquaTrackr</a>
	                    </div>
	                </nav>
	                <div class="container">
	        `)
		recent := showers.s[len(showers.s)-1]
		fmt.Fprintf(w, `
	            <div class="row">
                    <h1>%f Gallons</h1>
                </div>
	        `, recent.calcWater())
		fmt.Fprintln(w, `
                    </div>
                </body>
            `)
	})
	http.ListenAndServe(":8080", nil)
}
