package main

import (
	"fmt"
	"math/rand"
	"sync"
	
)

//-------------------------------------------STRUCTS-------------------------------------------
var wg sync.WaitGroup
var wg2 sync.WaitGroup

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




// ---------------------------------------------ZENTRALE STEUERLOGIK----------------------------------------
func ZentraleSteuerlogik() {

// woher bekommt die zentrale steuerlogik die aufzugalgorithmen?

	//slice erzeugen für auswertungen
	
	auswertungsListe := make([]Auswertung,0)
	
	anzahlSimulationsläufe := 3
	anzahlAufzüge := 3
	dauer := 10
	maxPersonen := 10
	
	// angegebene anzahl von Simulationsläufen aufrufen, auswertung wird als rückgabewert erhalten und an liste angehängt
	for i := 0; anzahlSimulationsläufe > i; i++{
		wg.Add(1)//DEADLOCK
		auswertung := Steuersimulation(anzahlAufzüge,dauer,maxPersonen)
		auswertungsListe = append(auswertungsListe,auswertung)
		fmt.Println("Auswertung:", auswertungsListe)
	}

	wg.Done()
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
	fmt.Println("steuersimulation vor empfange nachrichten generiege Aufzüge")
	
	for i:=0;i<anz;i++{
		neuerAufzug := <-channelGA
		fmt.Println("neuer Aufzug: ", neuerAufzug)
		aufzugListe = append(aufzugListe, neuerAufzug)	
		channelAlgA <- aufzugListe
	}
	fmt.Println("steuersimulation NACH empfange nachrichten generiege aufzüge")
	fmt.Println("aufzugliste:", aufzugListe)
    

		// erzeuge neue Personen
			
GenerierePassagiere(maxPers, channelGP)// übergibt berechnete maximal erlaubte personenanzahl
			
			//empfange Nachrichten von GenerierePassagiere
			fmt.Println("1000 steuersimulation vor empfange nachrichten generiege Passagiere")
			for i:=0;i<maxPers;i++{
			neueAnfragen := <-channelGP // speichere Nachricht in variabel
			fahrgaesteListe = append(fahrgaesteListe, neueAnfragen)// hänge neue anfrage an fahrgästeliste an
			channelAlgP <- fahrgaesteListe
			fmt.Println("fahrgästeliste: ", fahrgaesteListe)
			}	
			
			fmt.Println("steuersimulation NACH empfange nachrichten generiege Passagiere")
			

	wg2.Add(1)
	//starte Simulation
	 // überwacht gesamtlänge der simulation
		
		fmt.Println("steuersimulation:for dauer")

		fmt.Println("VOR Algorithmus")
		// algorithmus aufrufen
		go Aufzugsteuerungs_Agorithmus_1(channelAlgA, channelAlgP, maxPers, dauer, anz)			

		fmt.Println("NACH Algorithmus")
		// auswertung senden	
		
	
	select{
		case msg1 := <- channelAlgA:aufzugListe = msg1
		case msg2 := <- channelAlgP: fahrgaesteListe = msg2
		default:
		}
	//nach beendigung der simulation
	// erstelle auswertung	
		for i := range aufzugListe{
			auswertung.streckeA += aufzugListe[i].strecke // gesamtstrecke aller Aufzüge addiert
		}
		for i := range fahrgaesteListe{
			auswertung.wartezeitP += fahrgaesteListe[i].wartezeit //gesamtwartezeit der passagiere in schritten
			auswertung.abweichungP += fahrgaesteListe[i].abweichung // gesamtdifferenz von soll- zu ist-strecken der passagiere in schritten
		}

	fmt.Println("HIIIIIIIIIEEEEEEEEEEERRRRRR",fahrgaesteListe)
	
	wg2.Wait()
	wg.Done() // DEADLOCK
	return // gibt auswertung des simulationslaufs zurück
	
}

//--------------------------------------------GENERIERE AUFZÜGE---------------------------------------------------
func GeneriereAufzuege (anzA int, channelGA chan Aufzug){

	maxPers := 1 // wieviele dürfen höchstens mitfahren
	fmt.Println("for: anz Aufzüge = äußere schleife:", anzA)
	for i := 1; anzA >= i; i++{
		
			
			fmt.Println("AufzugNr:",i)
			
			neuerAufzug := Aufzug{aufzugNr: i, max: maxPers, event: "aufzug bereit"} 
			fmt.Println("neuer Aufzug",neuerAufzug)
			
			channelGA <- neuerAufzug
	}
	
}




//------------------------------------------GENERIERE PASSAGIERE-----------------------------------------------------

func GenerierePassagiere(max int, channelGP chan Person){ // hier kann anzahl an personen je schritt und anzahl etagen geändert werden
	fmt.Println("Einstieg generiere Passagiere")
	for ; max > 0; max--{
		
		wg2.Add(1)
		fmt.Println("for: anz Personen = äußere schleife:", max)
			startEtage := rand.Intn(4)
			zielEtage := 0 
			fmt.Println("startetage: ",startEtage,"zieletage:",zielEtage)
			
			for ungleich := false; ungleich == false;{ // damit start und zieletage nicht gleich sind
				zielEtage = rand.Intn(4)
				fmt.Println("zieletage", zielEtage)
				if startEtage != zielEtage{
					ungleich = true
				}
			}
			neueAnfrage := Person{start: startEtage, ziel: zielEtage} // wie macht man einen channel ????????
			fmt.Println("Generiere Person",neueAnfrage)
			channelGP <- neueAnfrage
				
			
	}
	

}






//---------------------------------------GOROUTINE PERSON-----------------------------------------------

func goroutineP(channelAlgPerson, channelAlgPerson2 chan Person, channelAlgAufzug, channelAlgAufzug2 chan Aufzug, maxpers, anz, dauer int ){
	
	aufzug := <- channelAlgAufzug
	fahrgast := <- channelAlgPerson

	select{ //for i:=0;i<anz*dauer;i++{
		case msg1 := <- channelAlgAufzug: aufzug = msg1
		case msg2 := <- channelAlgPerson: fahrgast = msg2
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
		wg2.Done()
	default: 
		
	}
	channelAlgAufzug2 <- aufzug
	channelAlgPerson2 <- fahrgast 
	

}

//------------------------------------------------------GOROUTINE AUFZUG------------------------------------------------------

func goroutineA(channelAlgPerson, channelAlgPerson2 chan Person, channelAlgAufzug, channelAlgAufzug2 chan Aufzug){

	aufzug := <- channelAlgAufzug
	fahrgast := <- channelAlgPerson

	select{ //for i:=0;i<anz*dauer;i++{
		case msg1 := <- channelAlgAufzug: aufzug = msg1
		case msg2 := <- channelAlgPerson: fahrgast = msg2
	}
	
	
	switch aufzug.event {
	case "aufzug bereit": 
		aufzug.zielEtage = fahrgast.start // aufzug erhält neues ziel

	case "aussteigen": 
		aufzug.fahrgaeste -= 1// anzahl der fahrgäste im Aufzug reduzieren

	case "einsteigen":
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


}


//----------------------------------------------ALGORITHMUS------------------------------------------------------------------

func Aufzugsteuerungs_Agorithmus_1 (channelAlgA chan []Aufzug, channelAlgP chan []Person, maxPers, dauer, anz int){
	// bekommt anfragen von personen eine je person?
	// steuert wege der Aufzüge
	

	fmt.Println("ALGORITHMUS EINSTIEG")
	aufzugListe := make([]Aufzug,0)
	fahrgaesteListe := make([]Person,0)
	chanfahrgast := make(chan Person,maxPers)
	chanaufzug := make(chan Aufzug,100)
	chan_antwort_aufzug := make(chan Aufzug,100)
	chan_antwort_fahrgast := make(chan Person,maxPers)

	aufzugListe = <- channelAlgA
	fahrgaesteListe = <- channelAlgP 

	go goroutineP(chanfahrgast,chan_antwort_fahrgast,chanaufzug,chan_antwort_aufzug, maxPers, anz, dauer)
	go goroutineA(chanfahrgast,chan_antwort_fahrgast,chanaufzug,chan_antwort_aufzug)
	
	for ; dauer >= 0; dauer--{
		
		select{
		case msg1 := <- chan_antwort_aufzug: aufzugListe = append(aufzugListe,msg1)
		}

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
			
					fmt.Println("AAAAAAAAAAAAAAAAA")
				
					break

				// schritt: einsteigen lassen
				}else if aufzugListe[i].etage == fahrgaesteListe[j].start && fahrgaesteListe[j].status == 1 && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr{
				
					aufzugListe[i].event = "einsteigen"
					fahrgaesteListe[j].status = 12
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
				
				
					break

				//schritt: zu anfrage runter fahren
				}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug runter fahren
				
					aufzugListe[i].event = "anfrage unten"
					//fahrgaesteListe[j].status = 1
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					
			
					break

				//schritt: zu anfrage rauf fahren
				}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 1{ // aufzug rauf fahren
					aufzugListe[i].event = "anfrage oben"
					//fahrgaesteListe[j].status = 1
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					
				
					break
				//schritt: zu ziel des gasts runter fahren
				}else if aufzugListe[i].etage > aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
					
					aufzugListe[i].event = "ziel unten"
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					
					break
				//schritt: zu ziel des gasts rauf fahren
				}else if aufzugListe[i].etage < aufzugListe[i].zielEtage && aufzugListe[i].aufzugNr == fahrgaesteListe[j].aufzugNr && fahrgaesteListe[j].status == 2{
					aufzugListe[i].event = "ziel oben"
					chanfahrgast <- fahrgaesteListe[j]
					chanaufzug <- aufzugListe[i]
					
				
					break
				}

			
			}
		
		}
		

		// ???????????????????????????????????????????????????????????????????????????????????????????
		fmt.Println("ALGORTHMUS Fahrgästeliste: ", fahrgaesteListe)
		fmt.Println("ALGORTHMUS Aufzugliste: ", aufzugListe)
		if dauer == 0{
			wg2.Done()
		}else{
			
			for j := range fahrgaesteListe{
				if fahrgaesteListe[j].status != 3{ // wenn noch nicht alle angekommen sind, abbrechen
				
					break
				}
				if j == len(fahrgaesteListe) - 1{
					wg2.Done()			// wenn die schleife bis hier hin gelaufen ist, sind alle angekommen
					break
				}
			
			

			}
		}

		

		channelAlgP <- fahrgaesteListe
		channelAlgA <- aufzugListe
		
	}
}
	
//}


//----------------------------------------------------MAIN--------------------------------------------------
func main(){	
	go ZentraleSteuerlogik()
	wg.Add(1) // wartet auf ein done --> mehr dones = crash
	
	wg.Wait()
	

	
}
