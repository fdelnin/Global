# CAAL

This are the propeties that can be prooved using CCS and the tool [CAAL](http://caal.cs.aau.dk/). There are two versions of the problem, one aims to prove that turns between students and the Dean of Students do not overlap while the other aims to prove some properties that affect the numebr of students in the room and the status of the Dean.

Compared to the original solution these two versions add some fake channels in order to comunicate with the external environment the status of the system, without these channels it's not possible to verify properties as the system performs only 'tau' steps because it only synchronizes internally.
