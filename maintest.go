package main

import (
	"math/rand"
	"sync"
	"fmt"
)

//-------------------------------------------STRUCTS-------------------------------------------
var wg sync.WaitGroup
var wg2 sync.WaitGroup


var wgSimulationsEnde sync.WaitGroup
var warteAufSteuerlogik sync.WaitGroup

type Aufzug struct {
	aufzugNr int // aufzug Nr
	fahrgaeste int	//Anzahl Passagiere
	max int  		//Maximalkapazität
	strecke int 	//Gesamt-Wegstrecke
	etage int // aktuelle Etage
	zielEtage int // nächster halt
	event string // 0 bereit, 1 einsteigend, 2 eingestiegen,3 fahrend, 4 aussteigend, 5 ausgestiegen
}
type Person struct {
	wartezeit int//Wartezeit in schritten
	soll int//Distanz start- ziel etage
	ist int// tatstächlich gefahrene Strecke
	abweichung int // differenz zwischen ist und soll
	start int // Startegage
	ziel int //Zieletage
	status int //  0 = wartend auf zuteilung, 1 = zugeteilt, 12 einsteigend, 2 = fahrend, 23 aussteigend, 3 = ausgestiegen
	aufzugNr int // nr des befördernden aufzugs
}
type Zentrale struct {
	anzSim int // anzahl an Simulationsläufen
	anzAufz int // anzahl Aufzüge
	maxPers int // maximale anzahl an personen pro simulation
	dauer int // dauer des simulationslaufs in Gesamtschritten
	algNr int // aufzugalgorithmen (Nr)
	auswertung []Auswertung // auswertung weiteres struct ???
}
type Auswertung struct { // alle aufzüge addiert!!!
	wartezeitP int //gesamtwartezeit der passagiere in schritten
	abweichungP int // gesamtdifferenz von soll- zu ist-strecken der passagiere in schritten
	streckeA int // gesamtstrecke aller Aufzüge addiert
}



var channelAuswertungAufzuege= make (chan []Aufzug,1)
var channelAuswertungPersonen=make(chan []Person,1)


// ---------------------------------------------ZENTRALE STEUERLOGIK----------------------------------------
func ZentraleSteuerlogik() {

	auswertungsListe := make([]Auswertung,0)

	anzahlSimulationsläufe := 1
	anzahlAufzüge := 4
	dauer := 100
	maxPersonen := 10

	// angegebene anzahl von Simulationsläufen aufrufen, auswertung wird als rückgabewert erhalten und an liste angehängt
	for i := 0; anzahlSimulationsläufe > i; i++{

		auswertung :=  Steuersimulation(anzahlAufzüge,dauer,maxPersonen)
		auswertungsListe = append(auswertungsListe,auswertung)

		println("Abweichung Personen",auswertung.abweichungP,"Strecke Aufzug=",auswertung.streckeA,"Wartezeit P=",auswertung.wartezeitP)
	}

	wgSimulationsEnde.Done() //Hier angekommen, sind die Simulationenfertig

}




//----------------------------------------------STEUERLOGIK EINES SIMULATIONSLAUFS---------------------------------------------------

func Steuersimulation (anz, dauer, maxPers int/*, algorithmus func()*/)(auswertung Auswertung){

	// slices erzeugen
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0)

	// channels erzeugen
	channelGP := make (chan Person,maxPers)
	channelGA := make (chan Aufzug,anz)
	channelAlgP := make (chan []Person, maxPers*4)
	channelAlgA := make (chan []Aufzug,anz*4)

	//generiere Aufzüge
	GeneriereAufzuege(anz, channelGA) // aufzüge erzeugen
	//empfange Nachrichten von generiere Aufzug, solange welche gesendet werden

	for i:=0;i<anz;i++{
		neuerAufzug := <-channelGA
		aufzugListe = append(aufzugListe, neuerAufzug)
			
	}
	channelAlgA <- aufzugListe
	// erzeuge neue Personen
	GenerierePassagiere(maxPers, channelGP)// übergibt berechnete maximal erlaubte personenanzahl
	//empfange Nachrichten von GenerierePassagiere
	for i:=0;i<maxPers;i++{
		neueAnfragen := <-channelGP // speichere Nachricht in variabel
		fahrgaesteListe = append(fahrgaesteListe, neueAnfragen)// hänge neue anfrage an fahrgästeliste an
		
	}
	channelAlgP <- fahrgaesteListe

	//starte Simulation
	// überwacht gesamtlänge der simulation

	// algorithmus aufrufen
	go Aufzugsteuerungs_Agorithmus_1(channelAlgA,channelAuswertungAufzuege, channelAlgP,channelAuswertungPersonen, maxPers, dauer, anz)
	// auswertung senden

	//Warte auf Aufzugssteuerung
	aufzugListe = <- channelAuswertungAufzuege
	fahrgaesteListe = <- channelAuswertungPersonen



	//nach beendigung der simulation
	// erstelle auswertung
	for i := range aufzugListe{
		auswertung.streckeA += aufzugListe[i].strecke // gesamtstrecke aller Aufzüge addiert
	}
	for i := range fahrgaesteListe{
		auswertung.wartezeitP += fahrgaesteListe[i].wartezeit //gesamtwartezeit der passagiere in schritten
		auswertung.abweichungP += fahrgaesteListe[i].abweichung // gesamtdifferenz von soll- zu ist-strecken der passagiere in schritten
	}


	//println("Abweichung Personen",auswertung.abweichungP,"Strecke Aufzug=",auswertung.streckeA,"Wartezeit P=",auswertung.wartezeitP)
	println("Eine Runde ist fertig")
	return // gibt auswertung des simulationslaufs zurück

}

//--------------------------------------------GENERIERE AUFZÜGE---------------------------------------------------
func GeneriereAufzuege (anzA int, channelGA chan Aufzug){

	maxPers := 1 // wieviele dürfen höchstens mitfahren
	for i := 1; anzA >= i; i++{


		neuerAufzug := Aufzug{aufzugNr: i, max: maxPers, event: "aufzug bereit"}

		channelGA <- neuerAufzug
	}

}




//------------------------------------------GENERIERE PASSAGIERE-----------------------------------------------------

func GenerierePassagiere(max int, channelGP chan Person){ // hier kann anzahl an personen je schritt und anzahl etagen geändert werden

	for ; max > 0; max--{

		startEtage := rand.Intn(4)
		zielEtage := 0


		for ungleich := false; ungleich == false;{ // damit start und zieletage nicht gleich sind
			zielEtage = rand.Intn(4)

			if startEtage != zielEtage{
				ungleich = true
			}
		}
		neueAnfrage := Person{start: startEtage, ziel: zielEtage} // wie macht man einen channel ????????

		channelGP <- neueAnfrage


	}


}






//---------------------------------------GOROUTINE PERSON-----------------------------------------------

func goroutineP(channelAlgPerson, channelAlgPerson2 chan Person, channelAlgAufzug, channelAlgAufzug2 chan Aufzug,doneChan chan bool){

	simuLaueft:=true
	for simuLaueft{

		var aufzug Aufzug
		var fahrgast Person

		select{ //for i:=0;i<anz*dauer;i++{
		case msg1 := <- channelAlgAufzug:
			aufzug = msg1
		default:
		}
		select{ //for i:=0;i<anz*dauer;i++{
		case msg2 := <- channelAlgPerson:
			fahrgast = msg2
		default:
		}

		switch fahrgast.status {
		case 0: fahrgast.ist += 1
			fahrgast.wartezeit += 1
		case 11:
			if aufzug.event == "aufzug bereit"{
				fahrgast.status = 1 // fahrgast hat jetzt einen zielaufzug
				fahrgast.aufzugNr = aufzug.aufzugNr
				fahrgast.soll = fahrgast.ziel - fahrgast.start
				if fahrgast.soll < 0{
					fahrgast.soll = fahrgast.soll *-1 // negative werte umkehren
				}
			}else{
				fahrgast.ist += 1
				fahrgast.wartezeit += 1
			}
		case 1: fahrgast.ist += 1
			fahrgast.wartezeit += 1
		case 12:
			fahrgast.status = 2
		case 2: fahrgast.ist += 1
		case 23:
			fahrgast.status = 3 // fahrgast als ausgestiegen markieren
			fahrgast.abweichung = fahrgast.ist - fahrgast.soll

		default:

		}
		channelAlgAufzug2 <- aufzug
		channelAlgPerson2 <- fahrgast

		//Stoppe  Schleife
		select {
		case msg1 := <-doneChan:
			simuLaueft = msg1
		default:
		}
	}


}

//------------------------------------------------------GOROUTINE AUFZUG------------------------------------------------------

func goroutineA(channelAlgPerson, channelAlgPerson2 chan Person, channelAlgAufzug, channelAlgAufzug2 chan Aufzug, doneChan chan bool){

	simuLaueft:=true
	for simuLaueft{
		//aufzug := <- channelAlgAufzug
		//fahrgast := <- channelAlgPerson
		var aufzug Aufzug
		var fahrgast Person
		select{ //for i:=0;i<anz*dauer;i++{
		case msg1 := <- channelAlgAufzug:
			aufzug = msg1
		default:
		}
		select{
		case msg2 := <- channelAlgPerson: fahrgast = msg2
		default:
		}


		switch aufzug.event {
		case "aufzug bereit":
			aufzug.zielEtage = fahrgast.start // aufzug erhält neues ziel
		case "aussteigen":
			//println("aussteigen")
			aufzug.fahrgaeste -= 1// anzahl der fahrgäste im Aufzug reduzieren
		case "einsteigen":
			//println("einsteigen")
			aufzug.fahrgaeste += 1
		case "anfrage unten":
			aufzug.etage -= 1
			aufzug.strecke += 1 // gesamtstrecke des aufzugs erhöhen
		case "anfrage oben":
			aufzug.etage += 1
			aufzug.strecke += 1 // gesamtstrecke des aufzugs erhöhen
		case "ziel unten":
			aufzug.etage -= 1
			aufzug.strecke += 1
		case "ziel oben":
			aufzug.etage += 1
			aufzug.strecke += 1
		}
		channelAlgAufzug2 <- aufzug
		channelAlgPerson2 <- fahrgast

		select { //for i:=0;i<anz*dauer;i++{
		case msg1 := <-doneChan:
			simuLaueft = msg1
		default:
		}
	}



}


//----------------------------------------------ALGORITHMUS------------------------------------------------------------------

func Aufzugsteuerungs_Agorithmus_1 (channelAlgA, channelAuswertungAufzuege chan []Aufzug, channelAlgP,channelAuswertungPersonen chan []Person, maxPers, dauer, anz int){
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0)


	chanfahrgast := make(chan Person,maxPers)
	chanaufzug := make(chan Aufzug,100)
	chan_antwort_aufzug := make(chan Aufzug,100)
	chan_antwort_fahrgast := make(chan Person,maxPers)

	//Channels um Routinen zu stoppen
	chanDoneAuf:=make(chan bool,1)
	chanDonePerRoutine:=make(chan bool,1)
	/*
	for ;maxPers > 0; maxPers--{
		select{
			case msg := <- channelAlgP: fahrgaesteListe = msg
			default:
		}

		//
	}
	for ;anz > 0; anz --{
		select{
			case msg := <- channelAlgA: aufzugListe = msg
			default:
			}
			//
	
	}
	*/
	aufzugListe = <- channelAlgA
	fahrgaesteListe = <- channelAlgP

	fmt.Println(aufzugListe)
	fmt.Println(fahrgaesteListe)
	go goroutineP(chanfahrgast,chan_antwort_fahrgast,chanaufzug,chan_antwort_aufzug,chanDonePerRoutine)
	go goroutineA(chanfahrgast,chan_antwort_fahrgast,chanaufzug,chan_antwort_aufzug,chanDoneAuf)

	for ; dauer >= 0; dauer--{

		antwort_fahrgast := <- chan_antwort_fahrgast
		antwort_aufzug := <- chan_antwort_aufzug
		fahrgaesteListe = append(fahrgaesteListe, antwort_fahrgast)
		aufzugListe = append(aufzugListe,antwort_aufzug)

		for i := range aufzugListe {
			// planung
			if aufzugListe[i].fahrgaeste == 0{
				//aufzug leer? dann erhalte neue anfrage
				for k:= range fahrgaesteListe{

					if fahrgaesteListe[k].status == 0{ // nehme ersten wartenden fahrgast

						aufzugListe[i].event = "aufzug bereit"
						fahrgaesteListe[k].status = 11
						chanfahrgast <- fahrgaesteListe[k]
						chanaufzug <- aufzugListe[i]
						fahrgaesteListe = append(fahrgaesteListe[:k], fahrgaesteListe[k:]...)
						aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)
						break
					}
				}
			}

			// ausführung
			for j:= range fahrgaesteListe{

				// schritt: aussteigen lassen
				if aufzugListe[i].etage == fahrgaesteListe[j].ziel && fahrgaesteListe[j].status == 2 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{

					aufzugListe[i].event = "aussteigen"
					fahrgaesteListe[j].status = 23
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j+1:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break

					// schritt: einsteigen lassen
				}else if aufzugListe[i].etage == fahrgaesteListe[j].start && fahrgaesteListe[j].status == 1 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{

					aufzugListe[i].event = "einsteigen"
					fahrgaesteListe[j].status = 12
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break

					//schritt: zu anfrage runter fahren
				}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug runter fahren

					aufzugListe[i].event = "anfrage unten"
					//fahrgaesteListe[j].status = 1
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break

					//schritt: zu anfrage rauf fahren
				}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug rauf fahren
					aufzugListe[i].event = "anfrage oben"
					//fahrgaesteListe[j].status = 1
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break
					//schritt: zu ziel des gasts runter fahren
				}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{

					aufzugListe[i].event = "ziel unten"
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break
					//schritt: zu ziel des gasts rauf fahren
				}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
					aufzugListe[i].event = "ziel oben"
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					fahrgaesteListe = append(fahrgaesteListe[:j], fahrgaesteListe[j:]...)
					aufzugListe = append(aufzugListe[:i], aufzugListe[i:]...)

					break
				}


			}

		}
		select{
		case msg1 := <- chan_antwort_fahrgast:
			fahrgaesteListe = append(fahrgaesteListe,msg1)
		case msg2 := <- chan_antwort_aufzug:
			aufzugListe = append(aufzugListe,msg2)
		default:
		}


	}

	chanDoneAuf<-false
	chanDonePerRoutine<-false
	//fmt.Println(aufzugListe)
	//fmt.Println(fahrgaesteListe)
	channelAuswertungPersonen <- fahrgaesteListe
	channelAuswertungAufzuege <- aufzugListe

}

//}


//----------------------------------------------------MAIN--------------------------------------------------
func main(){

	go ZentraleSteuerlogik()
	wgSimulationsEnde.Add(1) //Warte auf Simulationsende
	wgSimulationsEnde.Wait()
	println("Alles fertig")

}