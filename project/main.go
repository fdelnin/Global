package main

import "time"

const (
	PartyPeople int = 3  //number of people that makes a party
	Empty       int = 0  //number of people inside an "empty" room
	Students    int = 10 //Number of students
)

func Student(id int, turno chan bool, knock chan bool, permit chan bool, esci chan bool, changeTurnToDean chan bool) {
	var wait = true
	var myTurn bool
	var response bool //this is the answer to knocking
	println("Student n", id, "is in dormitory")

	for wait {
		myTurn = <-turno

		if myTurn == true {
			println("Student", id, "has been woken up")
			knock <- true
			response = <-permit
			if response { //sono entrato
				//passo il turno
				wait = false
				changeTurnToDean <- true
			} else {
				//nothing, aspetto e passo turno
				changeTurnToDean <- true
			}
		}

	}

	//party

	wait = true

	for wait {
		myTurn = <-turno
		if myTurn == true {
			println("Student", id, "has been woken up after partying")
			esci <- true
			wait = false
		}
	}

	println("Student", id, "is going home")
	changeTurnToDean <- true
	//se sono lultimo devo risvegliare ild ean
}

func Room(bussa chan bool, studentexits chan bool, askStatus chan bool, statusRoom chan string, checkdoor chan bool, answerpermit chan bool, entrato chan bool) {

	var numberOfStudents = 0
	var status = "empty"

	var permit bool

	for { //room never ends
		select {
		case <-bussa:

			//chiedi porta
			checkdoor <- true
			permit = <-answerpermit
			if permit {
				//se locked no se unlocked fai entrare
				if numberOfStudents == 0 {
					status = "someone"
				} else if numberOfStudents == PartyPeople-1 {
					status = "party"

				}
				numberOfStudents++
				println("studente ricevuto, status", status)
				entrato <- true

			} else {
				println("Studente non entra")
				entrato <- false
			}
		case <-studentexits:
			numberOfStudents--
			if numberOfStudents == 0 {
				status = "empty"
			} else if numberOfStudents == PartyPeople-1 {
				status = "someone"

			}
			println("studente uscito, status", status)

		case <-askStatus:
			println("status asked", status)
			statusRoom <- status

		}

	}

}

func Door(knocking chan bool, answer chan bool, lock chan bool, unlock chan bool) {

	var locked = false //unlocked at the beginning

	for { //does not end
		select {
		case <-knocking:
			if locked {
				println("door checked but room is locked")
				answer <- false
			} else {
				println("door checked and room is unlocked")
				answer <- true
			}

		case <-lock:
			if locked {
				println("Weird, room is already locked")
			}
			locked = true

		case <-unlock:
			if !locked {
				println("Weird, room is already unlocked")
			}
			locked = false

		}
	}

}

func Turn(changeToStudent chan bool, changeToDean chan bool, wakeDean chan bool, wakeAStudent chan bool) { //ask chan bool, answer chan bool,
	//	var currentTurn = true //true = turno studente

	for {
		select {
		//	case <-ask:
		//		answer <- currentTurn

		case <-changeToDean:
			//currentTurn = false
			println("Turn changed to dean")
			wakeDean <- true

		case <-changeToStudent:
			//	currentTurn = true
			println("Turn changed to student")
			wakeAStudent <- true
		}
	}

}

func Dean(askStatusRoom chan bool, answerStatusRoom chan string, lock chan bool, unlock chan bool, wakemedean chan bool, wakeAStudent chan bool) {

	var imInRoom = false
	//var myturn bool
	var room string
	var ImInterruptingParty = false //this is only to check if correct

	for {
		<-wakemedean //aspetat

		if imInRoom {
			askStatusRoom <- true
			room = <-answerStatusRoom
			if room == "empty" {
				imInRoom = false
				//unlock <- true
				if ImInterruptingParty {
					println("party is over")
					ImInterruptingParty = false

				} else {
					println("finished searching")
				}
				unlock <- true
				wakeAStudent <- true

			} else if room == "someone" {
				if ImInterruptingParty {
					println("I'm still waiting for students to exit party")
				} else {
					println("Error?")
				}

				wakeAStudent <- true

			} else { //room is party

				if ImInterruptingParty {
					println("I'm waiting for students to exit party")
				} else {
					println("Error party?")
				}
				wakeAStudent <- true
			}

		} else {
			askStatusRoom <- true
			room = <-answerStatusRoom
			if room == "empty" {
				lock <- true
				imInRoom = true
				println("started searching")
				wakeAStudent <- true

			} else if room == "someone" {
				wakeAStudent <- true

			} else { //room is party
				lock <- true
				imInRoom = true
				ImInterruptingParty = true
				println("interrupted party")
				wakeAStudent <- true
			}

		}

	}

}

func main() {
	knocking := make(chan bool)
	knockingAnswer := make(chan bool)
	doorAnswer := make(chan bool)
	//turn := make(chan bool)
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

	go Turn(changeTurnToStudent, changeTurnToDean, wakeDean, wakeAStudent)
	changeTurnToStudent <- true //initialize turn

	go Door(knocking, doorAnswer, lock, unlock)
	go Room(studentAskToEnter, studentexit, askStatusRoom, answerStatusRoom, knocking, doorAnswer, knockingAnswer)
	go Dean(askStatusRoom, answerStatusRoom, lock, unlock, wakeDean, changeTurnToStudent)

	var i = 0
	for i != Students {
		go Student(i, wakeAStudent, studentAskToEnter, knockingAnswer, studentexit, changeTurnToDean)
		i++
	}

	//to be removed:
	time.Sleep(30 * time.Second)
}
