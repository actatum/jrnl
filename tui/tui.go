package tui

import (
	"fmt"
	"log"
	"os"

	"github.com/actatum/jrnl"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the tui program.
func Run() error {
	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(basePath, os.ModePerm); err != nil {
		return err
	}

	if f, err := tea.LogToFile(basePath+"/debug.log", ""); err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	dbPath := basePath + "/db"
	jr, err := jrnl.NewJournal(dbPath)
	if err != nil {
		return err
	}
	defer func(jr *jrnl.Journal) {
		err := jr.Close()
		if err != nil {
			log.Printf("error closing journal: %v\n", err)
		}
	}(jr)

	initialized, err := jr.IsInitialized()
	if err != nil {
		return err
	}

	if initialized {
		// enter password prompt
		pw := EnterPasswordPrompt()
		err = jr.Auth(pw)
		if err != nil {
			return err
		}
	} else {
		// create password prompt
		pw, err := CreatePasswordPrompt()
		if err != nil {
			return err
		}

		fmt.Println("pw = ", pw)
		err = jr.CreatePassword(pw)
		if err != nil {
			return err
		}
	}

	_, err = jr.CreateEntry(`
[somelink](https://github.com)
Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Fusce ac ornare est, sed dapibus arcu. Vestibulum enim velit, blandit ac maximus eget, elementum at lacus. 
Nulla mollis placerat egestas. Nullam ullamcorper ex id efficitur ultrices. Sed et velit sapien. 
Proin dictum odio pharetra metus porta accumsan. 
Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. 
Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Suspendisse potenti.

Nunc nisi sem, tristique ut facilisis ut, ultricies at ante. 
Vivamus nec auctor leo. Praesent sit amet fermentum justo. 
Proin hendrerit augue quam, at vulputate ligula rutrum a. Mauris eu cursus justo. 
Sed vulputate ligula id orci sollicitudin, sed imperdiet ex eleifend. Etiam quis tortor nec ex viverra ullamcorper ut quis sem. 
Nam posuere posuere dolor vel hendrerit. Ut interdum scelerisque ipsum in maximus. In blandit fringilla pretium. 
Aenean efficitur et nibh non pellentesque. Donec congue neque vitae augue dignissim eleifend. 
Phasellus id ipsum quis neque lacinia fringilla. Morbi congue magna in felis facilisis, at bibendum nunc scelerisque.

Nulla vel erat bibendum, placerat quam at, convallis ante. Mauris sollicitudin vehicula consectetur. 
Sed bibendum justo diam, at tristique lacus gravida in. Nulla vitae purus et eros consequat ullamcorper. 
Pellentesque orci dui, tempor vitae lectus at, commodo mollis ligula. Proin pulvinar egestas tempus. 
In semper pharetra risus, vel pellentesque purus commodo a. Aenean hendrerit arcu lacus, sed suscipit est vestibulum vel. 
Aliquam sit amet odio mi. Suspendisse semper rutrum sem, non ullamcorper massa blandit vitae.

Pellentesque ac bibendum elit. Nulla quis tristique eros. Proin pretium sit amet tellus eu finibus. 
Duis tincidunt, nulla eu venenatis ullamcorper, lacus felis pharetra nisi, ac lobortis ligula neque maximus massa. 
Donec suscipit condimentum odio vitae facilisis. 
Duis ultricies, massa eget molestie convallis, mi urna dapibus velit, nec faucibus justo quam a massa. 
Nulla placerat, dolor condimentum consectetur ornare, augue orci mollis nibh, in egestas libero purus vel mauris. 
Mauris viverra leo sed dolor scelerisque, et commodo velit pellentesque. Sed sagittis sodales risus sit amet mattis. 
Donec sed suscipit enim. Morbi egestas, neque sed faucibus tristique, neque arcu vulputate nunc, placerat ultrices arcu massa at nisl. 
Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Integer in urna in est faucibus pretium.

Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. 
Proin sapien mi, tempus vel lectus a, aliquam lacinia augue. Duis ante mauris, dapibus in venenatis sed, volutpat ut nunc. 
Morbi consectetur odio arcu, eget condimentum diam commodo et. Pellentesque eget nisi mauris. Fusce at ante ante. 
Nunc eu vestibulum velit. Cras maximus sapien eu nulla molestie, ultricies feugiat ex congue. 
Praesent enim magna, sagittis quis nisl volutpat, viverra suscipit nisi. Aliquam erat volutpat. 
Sed ante dolor, pulvinar vitae mi sit amet, tempus pretium magna. `)
	if err != nil {
		return err
	}

	m, err := InitJournalUI(jr)
	if err != nil {
		return err
	}

	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return err
	}

	return nil
}
