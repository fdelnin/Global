package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

//Color for printing and number of people
const (
	StudentColor = "\033[1;34m%s\033[0m"
	DeanColor    = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"

	PartyPeople int = 5  //number of people that makes a party
	Students    int = 20 //Number of students
)

//Student function
func Student(id int, wakeMeUp chan bool, knock chan bool, permit chan bool, esci chan bool, studentSleep chan bool, waitForStudents chan bool) {

	var wait = true
	var response bool //this is the answer to knocking

	var printid string
	printid = strconv.Itoa(id)
	fmt.Printf("Student n " + printid + " is in dormitory\n")

	for wait {
		<-wakeMeUp

		var str = "Student " + printid + " has been woken up\n"
		fmt.Printf(StudentColor, str)
		knock <- true
		response = <-permit

		if response { //entered room
			wait = false
			str = "Student " + printid + " is inside the room\n"
			fmt.Printf(StudentColor, str)
			studentSleep <- true

		} else { //Can't enter the room

			str = "Student " + printid + " quietly walks away without enter\n"
			fmt.Printf(StudentColor, str)
			studentSleep <- true
		}

	}

	wait = true

	for wait {

		<-wakeMeUp
		var str = "Student " + printid + " has been woken up after partying, exits room\n"
		fmt.Printf(StudentColor, str)
		esci <- true
		wait = false

	}

	waitForStudents <- true

	fmt.Printf("Student %s is going home\n", printid)
}

//Room function
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
			if permit { //if unlocked

				if numberOfStudents == 0 {
					status = "someone"
				} else if numberOfStudents == PartyPeople-1 {
					status = "party"

				}
				numberOfStudents++
				fmt.Printf("Room status:%s\n", status)
				entrato <- true

			} else { //if locked no one can enter

				fmt.Printf("No one enter\n")
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
			statusRoom <- status

		case <-turnOffLight:
			b = false
		}

	}
	fmt.Printf("Lights are off, Room Thread is ended\n")
	closeDoorForNight <- true

}

//Door function
func Door(knocking chan bool, answer chan bool, doorLock chan bool,
	turnOffLight chan bool, closeDoorForNight chan bool) {
	var b = true
	var l bool
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

		case l = <-doorLock:

			if l { // true = lock
				if locked {
					fmt.Printf("Weird, room is already locked!\n")
				} else {
					fmt.Printf("Door is now locked\n")
				}
				locked = true

			} else { //false = unlock

				if !locked {
					fmt.Printf("Weird, room is already unlocked!\n")
				} else {
					fmt.Printf("Door is now unlocked\n")
				}
				locked = false
			}

		case <-closeDoorForNight:
			b = false
		}
	}

	fmt.Printf("Door Thread is ended\n")
	turnOffLight <- true

}

// Turn to change turns
func Turn(deanSleep chan bool, studentSleep chan bool, wakeDean chan bool, wakeAStudent chan bool) { //ask chan bool, answer chan bool,
	for {
		select {

		case <-studentSleep:

			var prob = randomProb()
			if prob >= 40 {
				fmt.Printf("\nDEAN TURN\n")
				wakeDean <- true
			} else {
				fmt.Printf("\nSTUDENT TURN\n")
				wakeAStudent <- true
			}

		case <-deanSleep:

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

//Dean function
func Dean(askStatusRoom chan bool, answerStatusRoom chan string, doorLock chan bool,
	wakemedean chan bool, endTurnDean chan bool, allGone chan bool, turnOffLight chan bool) {

	var imInRoom = false
	var room string
	var ImInterruptingParty = false //this is only to check if correct

	for {
		select {
		case <-wakemedean: //wake dean

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
					doorLock <- false //unlock door
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

					doorLock <- true //lock door
					imInRoom = true
					fmt.Printf(DeanColor, "Started searching inside the room\n")
					endTurnDean <- true

				} else if room == "someone" {

					fmt.Printf(DeanColor, "Waiting\n")
					endTurnDean <- true

				} else { //room is party

					fmt.Printf(DeanColor, "Party detected: I'm going to end this party\n")
					doorLock <- true //lock door
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

					doorLock <- false //unlock door
					fmt.Printf(DeanColor, "Dean exits room after all students are gone\n")

				} else {

					fmt.Printf(ErrorColor, "Error: some students are still inside nad Dean is going outside!")
				}
			}

			turnOffLight <- true //to stop the door and room threads
			<-turnOffLight
			fmt.Printf(DeanColor, "Dean goes home after all students\n")
			allGone <- true
		}
	}

}

func randomProb() int {
	// from 0 to 100
	return rand.Intn(100)

}

func main() {

	knocking := make(chan bool)
	knockingAnswer := make(chan bool)
	doorAnswer := make(chan bool)
	studentexit := make(chan bool)

	deanSleep := make(chan bool)
	studentSleep := make(chan bool)
	wakeDean := make(chan bool)
	wakeAStudent := make(chan bool)

	doorLock := make(chan bool)
	studentAskToEnter := make(chan bool)

	askStatusRoom := make(chan bool)
	answerStatusRoom := make(chan string)

	//this channels are for waiting functions to end
	turnOffLight := make(chan bool)
	closeDoorForNight := make(chan bool)
	waitForStudents := make(chan bool)
	allGone := make(chan bool)

	//get random problability to change turns
	rand.Seed(time.Now().UTC().UnixNano())

	go Turn(deanSleep, studentSleep, wakeDean, wakeAStudent)
	deanSleep <- true //initialize turn

	go Door(knocking, doorAnswer, doorLock, turnOffLight, closeDoorForNight)
	go Room(studentAskToEnter, studentexit, askStatusRoom, answerStatusRoom, knocking, doorAnswer, knockingAnswer, turnOffLight, closeDoorForNight)
	go Dean(askStatusRoom, answerStatusRoom, doorLock, wakeDean, deanSleep, allGone, turnOffLight)

	var i = 0
	for i != Students {
		go Student(i, wakeAStudent, studentAskToEnter, knockingAnswer, studentexit, studentSleep, waitForStudents)
		i++
	}

	var j = 0
	for j != Students {
		<-waitForStudents
		j++
		if j < Students {
			studentSleep <- true
		}
	}

	println("All students are gone\n")

	allGone <- false //tell dean everyone is gone, do not wait turn from student

	<-allGone //receive signal from dean

	println("END MAIN")
}
