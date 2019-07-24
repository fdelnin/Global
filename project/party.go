package main

import (
	"time"
)

func Porta(lock chan bool, unlock chan bool, status chan bool, answer chan bool, bussa chan bool, permit chan bool, entra chan bool) {
	var locked = false
	for {
		select {
		case <-lock:
			if locked == false {
				locked = true
				answer <- locked
				println("portachiusa")
			} else {
				println("errore porta già chiusa")
			}

		case <-unlock:
			if locked == true {
				locked = false
			} else {
				println("errore porta già aperta")
			}

		case <-status:
			answer <- !locked
			println("status delivered")

		case <-bussa:
			println("qualcuno ha bussato")
			if locked {
				println("no permit")
				permit <- false
			} else {
				println("si permit")
				permit <- true //avviso lo studente
				entra <- true
			}

		}

	}

}

func Turno(ask chan bool, answer chan bool, changeToStudent chan bool, changeToDean chan bool) {
	var turno = false //false student, true dean

	for {
		select {
		case <-ask:
			//println("turno", turno)
			answer <- turno

		case <-changeToDean:
			println("changing turn to dean")
			turno = true

		case <-changeToStudent:
			println("changing turn to student")
			turno = false

		}

	}
}

func Room(statusask chan bool, esci chan bool, entra chan bool, answer chan string) {
	var status string = "empty"
	var numerostudenti = 0

	for {
		select {
		case <-statusask:
			println("status", status)
			answer <- status

		case <-esci:

			if numerostudenti == 0 {
				println("errore 0 studenti")
			} else {
				numerostudenti -= 1
				if numerostudenti == 0 {
					status = "empty"
				} else if numerostudenti < 2 {
					status = "someone"
				}
			}

		case <-entra:
			numerostudenti += 1
			println("entra, ci sono ", numerostudenti, "studenti")

			if numerostudenti >= 2 {
				status = "party"
			} else {
				status = "someone"
			}

		}

	}
}

func Studente(bussa chan bool, permesso chan bool, entra chan bool, esci chan bool, askturno chan bool, answerturno chan bool, changetodean chan bool) {
	var b = true //entrato o meno
	var p bool   //permit
	var a bool   //answer my turn

	for b {
		askturno <- true
		a = <-answerturno
		if a == false {
			bussa <- true
			select {
			case p = <-permesso:
				if p == true {
					//entra <- true la porrta si apre da sola

					b = false
				} else {
					println("cedo il turno a dean perche non ho permesso")
					changetodean <- true
				}

			}
		} else {
			println("notmyturn")
			time.Sleep(2 * time.Second)
		}
	}
	changetodean <- true
	println("studente festeggia")
	var e = true
	for e {
		askturno <- true
		a = <-answerturno
		if a == false {
			esci <- true
			e = false
		} else {
			println("wait for uscire")
			time.Sleep(2 * time.Second)
		}

	}
	changetodean <- true
	println("studente va a casa")

}

func Dean(askturno chan bool, answerturno chan bool, lock chan bool, unlock chan bool, askroom chan bool, answerroom chan string,
	changeturntostudents chan bool, verifylocked chan bool) {
	var a bool //answer my turn
	var roomstatus string
	var iminroom = false
	var interrumptinparty = false //solo epr verifica che dean non entri mente ci ono meno di tot stud

	for {
		askturno <- true
		a = <-answerturno
		if a == false { //studenti
			//wait?
			println("not dean turn")
			time.Sleep(2 * time.Second)
		} else { //dean
			askroom <- true
			roomstatus = <-answerroom
			if iminroom && roomstatus == "someone" {
				if !interrumptinparty {
					println("EHM OPS")
				} else {
					println("STILl waiting guys")
				}
			}
			println("dean says room is ", roomstatus)
			if roomstatus == "someone" { //change turn cant do nothing
				changeturntostudents <- true

			} else if roomstatus == "empty" {
				if iminroom {
					unlock <- true
					println("deans unlocks room and exits")
					iminroom = false
					interrumptinparty = false
					changeturntostudents <- true
				} else {
					println("entrato e locked room")
					lock <- true
					<-verifylocked
					iminroom = true
					changeturntostudents <- true
				}

			} else if roomstatus == "party" {
				if !iminroom {
					println("dean detects party")
					lock <- true
					iminroom = true
					interrumptinparty = true
					changeturntostudents <- true
				} else {
					println("dean invite to hurry up")
					changeturntostudents <- true

				}

			}

		}

	}

}

func main() {
	bussare := make(chan bool)
	permesso := make(chan bool)
	lock := make(chan bool)
	unlock := make(chan bool)
	status := make(chan bool)
	statusanswer := make(chan bool)

	askturno := make(chan bool)
	answerturno := make(chan bool)
	changetostudent := make(chan bool)
	changetodean := make(chan bool)

	askroom := make(chan bool)
	entra := make(chan bool)
	esci := make(chan bool)
	roomstatus := make(chan string)

	go Porta(lock, unlock, status, statusanswer, bussare, permesso, entra)
	go Turno(askturno, answerturno, changetostudent, changetodean)
	go Room(askroom, esci, entra, roomstatus)
	go Studente(bussare, permesso, entra, esci, askturno, answerturno, changetodean)
	go Dean(askturno, answerturno, lock, unlock, askroom, roomstatus, changetostudent, statusanswer)
	time.Sleep(6 * time.Second)

	go Studente(bussare, permesso, entra, esci, askturno, answerturno, changetodean)
	go Studente(bussare, permesso, entra, esci, askturno, answerturno, changetodean)
	go Studente(bussare, permesso, entra, esci, askturno, answerturno, changetodean)
	time.Sleep(30 * time.Second)
}
