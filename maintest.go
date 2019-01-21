package main

import (
	"math/rand"
	"sync"
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

type aufzugUndPerson struct{
	p [] Person
	a [] Aufzug
}

type DatenPersonUndAuf struct{
	a Aufzug
	p Person
}

var channelAuswertungAufzuege= make (chan []Aufzug,1)
var channelAuswertungPersonen=make(chan []Person,1)


// ---------------------------------------------ZENTRALE STEUERLOGIK----------------------------------------
func ZentraleSteuerlogik() {

	auswertungsListe := make([]Auswertung,0)

	anzahlSimulationsläufe := 1
	anzahlAufzüge := 4
	dauer := 10000
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




	aufzugUndPersonCHAN:=make(chan aufzugUndPerson)



	//generiere Aufzüge
	GeneriereAufzuege(anz, channelGA) // aufzüge erzeugen
	//empfange Nachrichten von generiere Aufzug, solange welche gesendet werden

	for i:=0;i<anz;i++{
		neuerAufzug := <-channelGA
		aufzugListe = append(aufzugListe, neuerAufzug)

	}


	// erzeuge neue Personen
	GenerierePassagiere(maxPers, channelGP)// übergibt berechnete maximal erlaubte personenanzahl
	//empfange Nachrichten von GenerierePassagiere
	for i:=0;i<maxPers;i++{
		neueAnfragen := <-channelGP // speichere Nachricht in variabel
		fahrgaesteListe = append(fahrgaesteListe, neueAnfragen)// hänge neue anfrage an fahrgästeliste an

	}

	// algorithmus aufrufen
	go Aufzugsteuerungs_Agorithmus_1(aufzugUndPersonCHAN,maxPers, dauer, anz)
	// auswertung senden

	aufzugUndPersonCHAN<-aufzugUndPerson{fahrgaesteListe,aufzugListe}


	a:=<-aufzugUndPersonCHAN

	println("Test",len(a.p),len(a.a))

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


		neuerAufzug := Aufzug{i,0,maxPers,0,-3,-2,"aufzug bereit"}

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
		neueAnfrage := Person{0,0,0,0,startEtage,zielEtage	,0,0} // wie macht man einen channel ????????

		channelGP <- neueAnfrage


	}


}






//---------------------------------------GOROUTINE PERSON-----------------------------------------------

func goroutineP(chanMitAlgo chan DatenPersonUndAuf,doneChan chan bool){

	simuLaueft:=true
	for simuLaueft{

		var aufzug Aufzug
		var fahrgast Person

		neueDaten:=<-chanMitAlgo

		aufzug=neueDaten.a
		fahrgast= neueDaten.p
		switch fahrgast.status {
		case 0: fahrgast.ist += 1
			fahrgast.wartezeit += 1
		case 11:
			if aufzug.event == "aufzug bereit"{
				//println("Person hat Zielaufzug")
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
		case 1:
			fahrgast.ist += 1
			fahrgast.wartezeit += 1
			//println("Erhöhe um 1")
		case 12:
			fahrgast.status = 2
		case 2:
			fahrgast.ist += 1
		case 23:
			println("Person ist ausgestiegen markiert")
			fahrgast.status = 3 // fahrgast als ausgestiegen markieren
			fahrgast.abweichung = fahrgast.ist - fahrgast.soll
		default:

		}

		chanMitAlgo<-DatenPersonUndAuf{aufzug,fahrgast}

		//Stoppe  Schleife
		select {
		case msg1 := <-doneChan:
			simuLaueft = msg1
		default:
		}
	}


}

//------------------------------------------------------GOROUTINE AUFZUG------------------------------------------------------

func goroutineA(chanMitAlgorithmus chan DatenPersonUndAuf,done chan bool){


	//BEKOMMT AUFZUG UND UPDATET NACH EVENT
	//BEKOMMT PERSON ZIELETAGE
	simuLaueft:=true
	for simuLaueft{

		var aufzug Aufzug
		var fahrgast Person

		neueDaten:=<-chanMitAlgorithmus

		aufzug=neueDaten.a
		fahrgast= neueDaten.p
		switch aufzug.event {
		case "aufzug bereit":
			aufzug.zielEtage = fahrgast.start
			aufzug.fahrgaeste+=1	// aufzug erhält neues ziel
		case "aussteigen":
			//println("aussteigen")
			aufzug.fahrgaeste -= 1// anzahl der fahrgäste im Aufzug reduzieren
		case "einsteigen":
			aufzug.zielEtage=fahrgast.ziel
			//aufzug.fahrgaeste += 1
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

		chanMitAlgorithmus<- DatenPersonUndAuf{aufzug,fahrgast}

		select { //for i:=0;i<anz*dauer;i++{
		case msg1 := <-done:
			simuLaueft = msg1
			println("Fertig")
		default:
		}
	}



}


//----------------------------------------------ALGORITHMUS------------------------------------------------------------------

//func Aufzugsteuerungs_Agorithmus_1 (channelAlgA, channelAuswertungAufzuege chan []Aufzug, channelAlgP,channelAuswertungPersonen chan []Person, maxPers, dauer)


func Aufzugsteuerungs_Agorithmus_1 (aufzugUndPersonChan chan aufzugUndPerson,maxPers int, dauer int ,anz int){
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0)

	chanMitAufzug:=make(chan DatenPersonUndAuf)
	chanMitPerson:=make(chan DatenPersonUndAuf)


	chanDoneAuf:=make(chan bool,1)
	chanDonePerRoutine:=make(chan bool,1)


	aUp:=<-aufzugUndPersonChan

	aufzugListe =aUp.a
	fahrgaesteListe = aUp.p

	go goroutineP(chanMitPerson,chanDonePerRoutine)
	go goroutineA(chanMitAufzug,chanDoneAuf)


	for ; dauer >= 0; dauer--{
		for i := range aufzugListe {
			// planung
			if aufzugListe[i].fahrgaeste == 0{
				//aufzug leer? dann erhalte neue anfrage
				for k:= range fahrgaesteListe{
					if fahrgaesteListe[k].status == 0{ // nehme ersten wartenden fahrgast
						aufzugListe[i].event = "aufzug bereit"
						chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[k]}
						antwort:=<-chanMitAufzug
						aufzugListe[i]=antwort.a

						fahrgaesteListe[k].status = 11
						chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[k]}

						antwortPerson:=<-chanMitPerson
						fahrgaesteListe[k]=antwortPerson.p
						//println("Aufzug",aufzugListe[i].aufzugNr,"Aufzugfahrgäste",aufzugListe[i].fahrgaeste)
						break
					}
				}
			}

			// ausführung

			for j:= range fahrgaesteListe{
				//schritt: aussteigen lassen

				if aufzugListe[i].etage == fahrgaesteListe[j].ziel && fahrgaesteListe[j].status == 2 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{
					println("Jmd steigt Aus")
					aufzugListe[i].event = "aussteigen"
					fahrgaesteListe[j].status = 23
					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p
					// schritt: einsteigen lassen
				} else if aufzugListe[i].etage == fahrgaesteListe[j].start && fahrgaesteListe[j].status == 1 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{
					//println("Einsteigen")
					aufzugListe[i].event = "einsteigen"
					fahrgaesteListe[j].status = 12
					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p
					//println("Status",fahrgaesteListe[j].status)

					//schritt: zu anfrage runter fahren
				}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug runter fahren
					//println("Anfrage Unten")
					aufzugListe[i].event = "anfrage unten"

					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p


					//schritt: zu anfrage rauf fahren
				}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug rauf fahren

					//println("Anfrage Oben")
					aufzugListe[i].event = "anfrage oben"

					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p


					//schritt: zu ziel des gasts runter fahren
					} else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
						//println("Ziel Unten")
					aufzugListe[i].event = "ziel unten"

					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p



					//schritt: zu ziel des gasts rauf fahren
					}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
					//println("Ziel Oben")
					aufzugListe[i].event = "ziel oben"

					chanMitAufzug<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortAufzug:=<-chanMitAufzug
					chanMitPerson<-DatenPersonUndAuf{aufzugListe[i],fahrgaesteListe[j]}
					antwortPerson:=<-chanMitPerson
					aufzugListe[i]=antwortAufzug.a
					fahrgaesteListe[j]=antwortPerson.p

				}


			}

		}



	}

	chanDoneAuf<-false
	chanDonePerRoutine<-false
	aufzugUndPersonChan<-aufzugUndPerson{fahrgaesteListe,aufzugListe}

}




//----------------------------------------------------MAIN--------------------------------------------------
func main(){

	go ZentraleSteuerlogik()
	wgSimulationsEnde.Add(1) //Warte auf Simulationsende
	wgSimulationsEnde.Wait()
	println("Alles fertig")

}