package main

import (
	"../def"
	"../IO"
	"../ordermanager"
	"../fsm"
	//"net"
	//"../network/bcast"
	//"../network/localip"
	// "flag"
	"fmt"
	//"os"
	//"time"
)

func main() {
		backup := ordermanager.AmIBackup()
		fmt.Println("backup: ", backup)
	  IO.Init("localhost:15657", def.NUMFLOORS)

		ordermanager.InitElevMap(backup)
		go ordermanager.SoftwareBackup()

		var motor_direction IO.MotorDirection


		msg_buttonEvent := make(chan def.MapMessage, 100)
		msg_fromHWFloor := make(chan def.MapMessage, 100)
		msg_fromHWButton := make(chan def.MapMessage, 100)
		msg_toHW := make(chan def.MapMessage, 100)
		msg_toNetwork := make(chan def.MapMessage, 100)
		msg_fromNetwork := make(chan def.MapMessage, 100)
		msg_fromFSM := make(chan def.MapMessage, 100)
		msg_deadElev := make(chan def.MapMessage, 100)

	  drv_buttons := make(chan IO.ButtonEvent)
	  drv_floors  := make(chan int)
		fsm_chn			:= make(chan bool, 1)
		elevator_map_chn := make(chan def.MapMessage)

		go IO.PollButtons(drv_buttons)
		go IO.PollFloorSensor(drv_floors)


		drv_obstr   := make(chan bool)
	  drv_stop    := make(chan bool)
	  go IO.PollObstructionSwitch(drv_obstr)
	  go IO.PollStopButton(drv_stop)

		motor_direction = IO.MD_Down
		IO.SetMotorDirection(motor_direction)


		go fsm.FSM(drv_floors, fsm_chn, elevator_map_chn, motor_direction, msg_buttonEvent, msg_fromHWFloor, msg_fromHWButton, msg_fromFSM, msg_deadElev)

		for {
			select {
			case msg := <- msg_fromHWButton:
				msg_buttonEvent <- msg

			case msg := <- msg_fromNetwork:

			case msg := <- msg_fromFSM:

			default:

			}
		}



	    for {
					fmt.Println("Looping")
	        select {
	        case msg_button := <- drv_buttons:
	            fmt.Printf("%+v\n", msg_button)
	            IO.SetButtonLamp(msg_button.Button, msg_button.Floor, true)
							//bcast_chn <- msg_button

	        case msg_floor := <- drv_floors:
	            fmt.Printf("%+v\n", msg_floor)
	            if msg_floor == def.NUMFLOORS-1 {
	                motor_direction = IO.MD_Down
	            } else if msg_floor == 0 {
	                motor_direction = IO.MD_Up
	            }
	            IO.SetMotorDirection(motor_direction)


	        case msg_obstruction := <- drv_obstr:
	            fmt.Printf("%+v\n", msg_obstruction)
	            if msg_obstruction {
	                IO.SetMotorDirection(IO.MD_Stop)
	            } else {
	                IO.SetMotorDirection(motor_direction)
	            }

	        case msg_stop := <- drv_stop:
	            fmt.Printf("%+v\n", msg_stop)
	            for floor := 0; floor < def.NUMFLOORS; floor++ {
	                for button := IO.ButtonType(0); button < def.NUMBUTTON_TYPES; button++ {
	                    IO.SetButtonLamp(button, floor, false)
									}
							}

					//case msg_recieve := <- recieve_chn:
						//	fmt.Printf("%+v\n", msg_recieve)

	        }
	    }
	}
