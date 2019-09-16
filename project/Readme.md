# The Room Party

This is a problem from "The Little Book of Semaphores", Allen B. Downey, I propose two solutions to this problem: one implemented in `go` the other in CCS.

## CCS

This solution is infinite, all students want to enter the party, after some time they exit and then they start again, the Dean of Students also waits forever for the room to be empty or for a party to be interrupted.

Some properties can be proved with HML and CCS itself, using the tool [CAAL](http://caal.cs.aau.dk/).

## GO

This implementation is finite, all students want to party, after some time they exit the room and go home. The Dean of Students may be inside the room when the last student leave the dormitory, so a signal must be sent to the Dean to let him know his working day is over. Before leaving the building he also turns off the lights in the room.

The yellow prints are from the Dean of Students, when he/she performs an action inside the critical section, the blue ones are from the students, red ones should never be printed as they are logical errors with respect to the specification of the problem.
