package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const (
	StudentColor = "\033[1;34m%s\033[0m"
	DeanColor    = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"

	PartyPeople int = 5  //number of people that makes a party
	Students    int = 20 //Number of students
)

func Student(id int, turno chan bool, knock chan bool, permit chan bool, esci chan bool, studentSleep chan bool, waitForStudents chan bool) {
	var wait = true
	var myTurn bool
	var response bool //this is the answer to knocking

	var printid string
	printid = strconv.Itoa(id)
	fmt.Printf("Student n " + printid + " is in dormitory\n")

	for wait {
		myTurn = <-turno

		if myTurn == true {
			var str = "Student " + printid + " has been woken up\n"
			fmt.Printf(StudentColor, str)
			knock <- true
			response = <-permit
			if response { //sono entrato
				//passo il turno
				wait = false
				str = "Student " + printid + " is inside the room\n"
				fmt.Printf(StudentColor, str)
				studentSleep <- true
			} else {
				//non posso entrare, aspetto
				studentSleep <- true
			}
		}

	}

	wait = true

	for wait {
		myTurn = <-turno
		if myTurn == true {
			var str = "Student " + printid + " has been woken up after partying, exits room\n"
			fmt.Printf(StudentColor, str)
			esci <- true
			wait = false
		}
	}

	studentSleep <- true

	waitForStudents <- true

	fmt.Printf("Student %s is going home\n", printid)
}

func Room(knock chan bool, studentexits chan bool, askStatus chan bool, statusRoom chan string,
	checkdoor chan bool, answerpermit chan bool, entrato chan bool, turnOffLight chan bool, closeDoorForNight chan bool) {

	var b = true
	var numberOfStudents = 0
	var status = "empty"

	var permit bool

	for b { //room ends when receives a signal on turnOffLightsChannel
		select {
		case <-knock:
			//check if door is locked
			checkdoor <- true
			permit = <-answerpermit
			if permit {
				//if locked no one can enter
				if numberOfStudents == 0 {
					status = "someone"
				} else if numberOfStudents == PartyPeople-1 {
					status = "party"

				}
				numberOfStudents++
				fmt.Printf("Room status:%s\n", status)
				entrato <- true

			} else {
				fmt.Printf("A student quietly walks away without enter\n")
				entrato <- false
			}

		case <-studentexits:
			numberOfStudents--

			if numberOfStudents == 0 {
				status = "empty"
			} else if numberOfStudents == PartyPeople-1 {
				status = "someone"
			}
			fmt.Printf("Room status:%s\n", status)

		case <-askStatus:
			//println("status asked", status)
			statusRoom <- status

		case <-turnOffLight:
			b = false
		}

	}
	fmt.Printf("Lights are off, Room Thread is ended\n")
	closeDoorForNight <- true

}

func Door(knocking chan bool, answer chan bool, lock chan bool, unlock chan bool,
	turnOffLight chan bool, closeDoorForNight chan bool) {
	var b = true
	var locked = false //unlocked at the beginning

	for b { //ends only when lights in the room are turned off
		select {
		case <-knocking:
			if locked {
				fmt.Printf("Door checked but room is locked\n")
				answer <- false
			} else {
				fmt.Printf("Door checked and room is unlocked\n")
				answer <- true
			}

		case <-lock:
			if locked {
				fmt.Printf("Weird, room is already locked!\n")
			} else {
				fmt.Printf("Door is now locked\n")
			}
			locked = true

		case <-unlock:
			if !locked {
				fmt.Printf("Weird, room is already unlocked!\n")
			} else {
				fmt.Printf("Door is now unlocked\n")
			}
			locked = false

		case <-closeDoorForNight:
			b = false
		}
	}

	fmt.Printf("Door Thread is ended\n")
	turnOffLight <- true

}

func Turn(changeToStudent chan bool, studentWait chan bool, wakeDean chan bool, wakeAStudent chan bool) { //ask chan bool, answer chan bool,
	for {
		select {
		case <-studentWait:
			var prob = randomProb()
			if prob >= 40 {
				fmt.Printf("\nDEAN TURN\n")
				wakeDean <- true
			} else {
				fmt.Printf("\nSTUDENT TURN\n")
				wakeAStudent <- true
			}
		case <-changeToStudent:

			var prob = randomProb()
			if prob >= 60 {
				fmt.Printf("\nDEAN TURN\n")
				wakeDean <- true
			} else {
				fmt.Printf("\nSTUDENT TURN\n")
				wakeAStudent <- true
			}

		}
	}

}

func Dean(askStatusRoom chan bool, answerStatusRoom chan string, lock chan bool, unlock chan bool,
	wakemedean chan bool, endTurnDean chan bool, allGone chan bool, turnOffLight chan bool) {

	var imInRoom = false
	//var myturn bool
	var room string
	var ImInterruptingParty = false //this is only to check if correct

	for {
		select {
		case <-wakemedean: //aspetat
			askStatusRoom <- true
			room = <-answerStatusRoom

			if imInRoom {
				if room == "empty" {
					imInRoom = false
					if ImInterruptingParty {
						fmt.Printf(DeanColor, "Party is finally over\n")
						ImInterruptingParty = false

					} else {
						fmt.Printf(DeanColor, "Finished searching\n")
					}
					unlock <- true
					endTurnDean <- true

				} else if room == "someone" {
					if ImInterruptingParty {
						fmt.Printf(DeanColor, "I'm still waiting for students to exit party\n")
					} else {
						fmt.Printf(ErrorColor, "Error - Dean inside room with only some students")
					}

					endTurnDean <- true

				} else { //room is party

					if ImInterruptingParty {
						fmt.Printf(DeanColor, "I'm waiting for students to exit party\n")
					} else {
						fmt.Printf(ErrorColor, "Error - Dean inside but not interrupting party")
					}
					endTurnDean <- true
				}

			} else {

				if room == "empty" {
					lock <- true
					imInRoom = true
					fmt.Printf(DeanColor, "Started searching inside the room\n")
					endTurnDean <- true

				} else if room == "someone" {
					fmt.Printf(DeanColor, "Waiting\n")
					endTurnDean <- true

				} else { //room is party
					fmt.Printf(DeanColor, "Party detected: I'm going to end this party\n")
					lock <- true
					imInRoom = true
					ImInterruptingParty = true
					endTurnDean <- true
				}
			}

		case <-allGone:

			if imInRoom {
				askStatusRoom <- true
				room = <-answerStatusRoom
				if room == "empty" {
					unlock <- true
					fmt.Printf(DeanColor, "Dean exits room after all students are gone\n")
				} else {
					fmt.Printf(ErrorColor, "Error: some students are still inside nad Dean is going outside!")
				}
			}

			turnOffLight <- true //to stop the door and room threads
			<-turnOffLight
			fmt.Printf("Dean goes home after all students\n")
			allGone <- true
		}
	}

}

func randomProb() int {
	// from 0 to 100
	return rand.Intn(100)

}

func main() {
	// TODO: use values of channels instead of so many channels
	knocking := make(chan bool)
	knockingAnswer := make(chan bool)
	doorAnswer := make(chan bool)
	studentexit := make(chan bool)

	changeTurnToStudent := make(chan bool)
	changeTurnToDean := make(chan bool)
	wakeDean := make(chan bool)
	wakeAStudent := make(chan bool)

	lock := make(chan bool)
	unlock := make(chan bool)
	studentAskToEnter := make(chan bool)

	askStatusRoom := make(chan bool)
	answerStatusRoom := make(chan string)
	waitForStudents := make(chan bool)
	allGone := make(chan bool)

	//this channels are for waiting functions end
	turnOffLight := make(chan bool)
	closeDoorForNight := make(chan bool)
	//get random problability to change turns
	rand.Seed(time.Now().UTC().UnixNano())

	go Turn(changeTurnToStudent, changeTurnToDean, wakeDean, wakeAStudent)
	changeTurnToStudent <- true //initialize turn

	go Door(knocking, doorAnswer, lock, unlock, turnOffLight, closeDoorForNight)
	go Room(studentAskToEnter, studentexit, askStatusRoom, answerStatusRoom, knocking, doorAnswer, knockingAnswer, turnOffLight, closeDoorForNight)
	go Dean(askStatusRoom, answerStatusRoom, lock, unlock, wakeDean, changeTurnToStudent, allGone, turnOffLight)

	var i = 0
	for i != Students {
		go Student(i, wakeAStudent, studentAskToEnter, knockingAnswer, studentexit, changeTurnToDean, waitForStudents)
		i++
	}

	var j = 0
	for j != Students {
		<-waitForStudents
		j++
	}

	println("All students are gone\n")

	allGone <- false //tell dean everyone is gone, do not wait turn from student

	<-allGone //receive signal from dean

	println("END MAIN")
}
