package main

var wallboxStatusCodes = []string{
	"Ready",
	"Charging",
	"Connected waiting car",
	"Connected waiting schedule",
	"Paused",
	"Schedule end",
	"Locked",
	"Error",
	"Connected waiting current assignation",
	"Unconfigured power sharing",
	"Queue by power boost",
	"Discharging",
	"Connected waiting admin auth for mid",
	"Connected mid safety margin exceeded",
	"OCPP unavailable",
	"OCPP charge finishing",
	"OCPP reserved",
	"Updating",
	"Queue by eco smart",
}

var stateOverrides = map[int]int{
	0xa1: 0,
	0xa2: 9,
	0xa3: 0xe,
	0xa4: 0xf,
	0xa6: 0x11,
	0xb1: 3,
	0xb3: 3,
	0xb2: 4,
	0xb6: 4,
	0xb4: 2,
	0xb5: 2,
	0xb7: 8,
	0xb8: 8,
	0xb9: 10,
	0xba: 10,
	0xbb: 0xc,
	0xbc: 0xd,
	0xbd: 0x12,
	0xc1: 1,
	0xc2: 1,
	0xc3: 0xb,
	0xc4: 0xb,
	0xd1: 6,
	0xd2: 6,
}

var stateMachineStates = map[int]string{
	0xE:  "Error",
	0xF:  "Unviable",
	0xA1: "Ready",
	0xA2: "PS Unconfig",
	0xA3: "Unavailable",
	0xA4: "Finish",
	0xA5: "Reserved",
	0xA6: "Updating",
	0xB1: "Connected 1", // Make new session?
	0xB2: "Connected 2",
	0xB3: "Connected 3", // Waiting schedule ?
	0xB4: "Connected 4",
	0xB5: "Connected 5", // Connected waiting car ?
	0xB6: "Connected 6", // Paused
	0xB7: "Waiting 1",
	0xB8: "Waiting 2",
	0xB9: "Waiting 3",
	0xBA: "Waiting 4",
	0xBB: "Mid 1",
	0xBC: "Mid 2",
	0xBD: "Waiting eco power",
	0xC1: "Charging 1",
	0xC2: "Charging 2",
	0xC3: "Discharging 1",
	0xC4: "Discharging 2",
	0xD1: "Lock",
	0xD2: "Wait Unlock",
}

var controlPilotStates = map[int]string{
	0xE:  "Error",
	0xF:  "Failure",
	0xA1: "Ready 1", // S1 at 12V, car not connected
	0xA2: "Ready 2",
	0xB1: "Connected 1", // S1 at 9V, car connected not allowed charge
	0xB2: "Connected 2", // S1 at Oscillator, car connected allowed charge
	0xC1: "Charging 1",
	0xC2: "Charging 2", // S2 closed
}
