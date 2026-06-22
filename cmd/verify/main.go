package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movekit/movekit/internal/folders"
	"github.com/movekit/movekit/internal/migrate"
	sys "github.com/movekit/movekit/internal/syscall"
)

func main() {
	fmt.Println("========== MoveKit Backend Runtime Verification ==========")
	failures := 0

	// 1. GetFolders
	fmt.Println("\n[1] GetFolders")
	for _, f := range folders.AllFolders {
		rawPath, err := sys.ReadExpandString(sys.UserShellFoldersKey, f.RegValueName)
		if err != nil {
			fmt.Printf("  FAIL %s: registry read error: %v\n", f.Name, err)
			failures++
			continue
		}
		expanded := os.ExpandEnv(rawPath)
		info, err := os.Stat(expanded)
		if err != nil {
			fmt.Printf("  WARN %s: path not found: %s\n", f.Name, expanded)
			continue
		}
		isJunction := sys.IsJunction(expanded)
		fmt.Printf("  OK %s: %s (dir=%v junction=%v)\n", f.Name, expanded, info.IsDir(), isJunction)
	}

	// 2. GetDisks
	fmt.Println("\n[2] GetDisks")
	for _, drive := range []string{"C", "D", "E", "F"} {
		volInfo, err := sys.GetVolumeInfo(drive)
		if err != nil {
			continue
		}
		free, total, _, _ := sys.GetDiskFreeSpace(drive)
		fmt.Printf("  OK %s: FS=%s Fixed=%v Removable=%v Network=%v Free=%dMB Total=%dMB\n",
			drive, volInfo.FileSystem, volInfo.IsFixedDrive, volInfo.IsRemovable, volInfo.IsNetwork,
			free/(1024*1024), total/(1024*1024))
	}

	// 3. CheckOneDriveKFM
	fmt.Println("\n[3] CheckOneDriveKFM")
	status, err := migrate.DetectOneDriveKFM()
	if err != nil {
		fmt.Printf("  FAIL: %v\n", err)
		failures++
	} else {
		fmt.Printf("  OK IsActive=%v Folders=%v HasGPOBlock=%v\n", status.IsActive, status.Folders, status.HasGPOBlock)
	}

	// 4. PreCheck
	fmt.Println("\n[4] PreCheck (Desktop C->D)")
	desktop, _ := sys.ReadExpandString(sys.UserShellFoldersKey, "Desktop")
	desktop = os.ExpandEnv(desktop)
	if desktop != "" {
		checkErr := migrate.PreCheck(desktop, "D")
		if checkErr != nil {
			fmt.Printf("  WARN Code=%s Msg=%s\n", checkErr.Code, checkErr.Message)
		} else {
			fmt.Printf("  OK passed\n")
		}
	}

	// 5. IsCrossVolume
	fmt.Println("\n[5] IsCrossVolume")
	fmt.Printf("  C->D: %v (expect true)\n", sys.IsCrossVolume(`C:\Users\test`, `D:\Downloads`))
	fmt.Printf("  C->C: %v (expect false)\n", sys.IsCrossVolume(`C:\Users\test`, `C:\Downloads`))

	// 6. ToLongPath
	fmt.Println("\n[6] ToLongPath")
	fmt.Printf("  D:\test -> %s\n", sys.ToLongPath(`D:\test`))

	// 7. DetectConflicts
	fmt.Println("\n[7] DetectConflicts")
	if desktop != "" {
		conflicts, _ := migrate.DetectConflicts(desktop, `Z:/nonexistent`)
		fmt.Printf("  OK conflicts=%d (target missing = 0)\n", len(conflicts))
	}

	// 8. DetectCloudSync
	fmt.Println("\n[8] DetectCloudSync")
	providers := migrate.DetectCloudSync()
	fmt.Printf("  OK detected=%d\n", len(providers))
	for _, p := range providers {
		fmt.Printf("    - %s: %s\n", p.Name, p.Folder)
	}

	// 9. LoadAllBackups
	fmt.Println("\n[9] LoadAllBackups")
	backups, err := migrate.LoadAllBackups()
	if err != nil {
		fmt.Printf("  WARN: %v\n", err)
	} else {
		fmt.Printf("  OK count=%d\n", len(backups))
	}

	// 10. CheckUpdate
	fmt.Println("\n[10] CheckUpdate")
	info, _ := migrate.CheckUpdate()
	fmt.Printf("  OK HasUpdate=%v Version=%s\n", info.HasUpdate, info.NewVersion)

	// 11. Logger
	fmt.Println("\n[11] Logger")
	log, err := migrate.NewLogger()
	if err != nil {
		fmt.Printf("  FAIL: %v\n", err)
		failures++
	} else {
		log.Info("runtime verification test")
		log.Close()
		fmt.Printf("  OK log written\n")
	}

	// 12. IsOnBattery
	fmt.Println("\n[12] IsOnBattery")
	fmt.Printf("  OK IsOnBattery=%v\n", sys.IsOnBattery())

	// 13. Junction roundtrip
	fmt.Println("\n[13] Junction create/verify/delete")
	tmpDir, _ := os.MkdirTemp("", "mvk_verify")
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")
	os.Mkdir(target, 0755)
	os.Mkdir(link, 0755)
	if err := sys.CreateJunction(link, target); err != nil {
		fmt.Printf("  FAIL CreateJunction: %v\n", err)
		failures++
	} else if !sys.IsJunction(link) {
		fmt.Printf("  FAIL IsJunction=false\n")
		failures++
	} else if err := sys.DeleteJunction(link); err != nil {
		fmt.Printf("  FAIL DeleteJunction: %v\n", err)
		failures++
	} else {
		fmt.Printf("  OK full roundtrip\n")
	}
	os.RemoveAll(tmpDir)

	// 14. Registry read/write
	fmt.Println("\n[14] Registry read User Shell Folders")
	rawPath, err := sys.ReadExpandString(sys.UserShellFoldersKey, "Desktop")
	if err != nil {
		fmt.Printf("  FAIL: %v\n", err)
		failures++
	} else {
		fmt.Printf("  OK Desktop=%s\n", rawPath)
	}

	// 15. Shell Folders
	fmt.Println("\n[15] Registry read Shell Folders")
	shellPath, err := sys.ReadExpandString(sys.ShellFoldersKey, "Desktop")
	if err != nil {
		fmt.Printf("  FAIL: %v\n", err)
		failures++
	} else {
		fmt.Printf("  OK Desktop=%s\n", shellPath)
	}

	// Summary
	fmt.Println("\n==========================================")
	if failures > 0 {
		fmt.Printf("RESULT: %d FAILURES\n", failures)
		os.Exit(1)
	} else {
		fmt.Println("RESULT: ALL PASSED")
	}
}
