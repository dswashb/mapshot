package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Palats/mapshot/embed"
	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RenderFlags holds parameters to the rendering.
type RenderFlags struct {
	tilemin    int64
	tilemax    int64
	prefix     string
	resolution int64
	jpgquality int64
}

// Register creates flags for the rendering parameters.
func (rf *RenderFlags) Register(flags *pflag.FlagSet, prefix string) *RenderFlags {
	flags.Int64Var(&rf.tilemin, prefix+"tilemin", 0, "Size in in-game units of a tile for the most zoomed layer. If 0, use value from the game.")
	flags.Int64Var(&rf.tilemax, prefix+"tilemax", 0, "Size in in-game units of a tile for the least zoomed layer. If 0, use value from the game.")
	flags.StringVar(&rf.prefix, prefix+"prefix", "", "Prefix to add to all generated filenames. If empty, use value from the game.")
	flags.Int64Var(&rf.resolution, prefix+"resolution", 0, "Pixel size for generated tiles. If 0, use value from the game.")
	flags.Int64Var(&rf.jpgquality, prefix+"jpgquality", 0, "Compression quality for jpg files. If 0, use value from the game.")
	return rf
}

func (rf *RenderFlags) genOverrides() map[string]interface{} {
	ov := map[string]interface{}{}
	if rf.tilemin != 0 {
		ov["tilemin"] = rf.tilemin
	}
	if rf.tilemax != 0 {
		ov["tilemax"] = rf.tilemax
	}
	if rf.prefix != "" {
		ov["prefix"] = rf.prefix
	}
	if rf.resolution != 0 {
		ov["resolution"] = rf.resolution
	}
	if rf.jpgquality != 0 {
		ov["jpqquality"] = rf.jpgquality
	}
	return ov
}

func copyMod(dstMapshot string) error {
	if err := os.MkdirAll(dstMapshot, 0755); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", dstMapshot, err)
	}
	for name, content := range embed.ModFiles {
		dst := path.Join(dstMapshot, name)
		if err := ioutil.WriteFile(dst, []byte(content), 0644); err != nil {
			return fmt.Errorf("unable to write file %q: %w", dst, err)
		}
	}
	return nil
}

func writeOverrides(data map[string]interface{}, dstPath string) error {
	inline, err := json.Marshal(data)
	if err != nil {
		return err
	}
	overrides := "return [===[" + string(inline) + "]===]\n"
	overridesFilename := path.Join(dstPath, "overrides.lua")
	if err := ioutil.WriteFile(overridesFilename, []byte(overrides), 0644); err != nil {
		return fmt.Errorf("unable to write overrides file %q: %w", overridesFilename, err)
	}
	glog.Infof("overrides file created at %q", overridesFilename)
	return nil
}

func render(ctx context.Context, factorioSettings *factorio.Settings, rf *RenderFlags, rawname string) error {
	fact, err := factorio.New(factorioSettings)
	if err != nil {
		return err
	}

	runID := uuid.New().String()
	glog.Infof("runid: %s", runID)

	// The parameter can be a filename, so extract a name.
	name := path.Base(rawname)
	name = name[:len(name)-len(path.Ext(name))]

	tmpdir, err := ioutil.TempDir("", "mapshot")
	if err != nil {
		return fmt.Errorf("unable to create temp dir: %w", err)
	}
	glog.Info("temp dir: ", tmpdir)

	// Copy game save
	srcSavegame, err := fact.FindSaveFile(rawname)
	if err != nil {
		return fmt.Errorf("unable to find savegame %q: %w", rawname, err)
	}
	fmt.Printf("Generating mapshot %q using file %s\n", name, srcSavegame)

	dstSavegame := path.Join(tmpdir, name+".zip")
	if err := copy.Copy(srcSavegame, dstSavegame); err != nil {
		return fmt.Errorf("unable to copy file %q: %w", srcSavegame, err)
	}
	glog.Infof("copied save from %q to %q", srcSavegame, dstSavegame)

	// Copy mods
	dstMods := path.Join(tmpdir, "mods")
	if err := fact.CopyMods(dstMods, []string{"mapshot"}); err != nil {
		return err
	}

	// Add the mod itself.
	dstMapshot := path.Join(dstMods, "mapshot")
	if err := copyMod(dstMapshot); err != nil {
		return err
	}
	if err := factorio.EnableMod(dstMods, "mapshot"); err != nil {
		return err
	}
	glog.Infof("mod created at %q", dstMapshot)

	// Generates overrides to the parameters. This is done by creating a Lua
	// file, as mods don't have any way of loading data.
	overridesData := rf.genOverrides()
	overridesData["onstartup"] = runID
	overridesData["shotname"] = name
	if err := writeOverrides(overridesData, dstMapshot); err != nil {
		return err
	}

	// Remove done marker if still present
	var doneFile string
	doneFile = path.Join(fact.ScriptOutput(), "mapshot-done-"+runID)
	err = os.Remove(doneFile)
	glog.Infof("removed done-file %q: %v", doneFile, err)

	factorioArgs := []string{
		"--disable-audio",
		"--disable-prototype-history",
		"--load-game", dstSavegame,
		"--mod-directory", dstMods,
	}
	glog.Infof("Factorio args: %v", factorioArgs)

	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error)
	fmt.Println("Starting Factorio...")
	go func() {
		errCh <- fact.Run(execCtx, factorioArgs)
	}()

	// Wait for the `done` file to be created, indicating that the work is
	// done.
	for {
		_, err := os.Stat(doneFile)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to stat file %q: %w", doneFile, err)
		}
		if err == nil {
			cancel()
			break
		}

		// Context cancellation should terminate Factorio, which is detected
		// through errCh, so no need to wait on context.
		select {
		case <-time.After(time.Second):
		case err := <-errCh:
			return fmt.Errorf("factorio exited early; err=%w", err)
		}
	}
	glog.Infof("done file %q now exists", doneFile)
	rawDone, err := ioutil.ReadFile(doneFile)
	if err != nil {
		return fmt.Errorf("unable to read file %q: %w", doneFile, err)
	}
	resultPrefix := string(rawDone)
	glog.Infof("output at %s", resultPrefix)
	fmt.Println("Output:", path.Join(fact.ScriptOutput(), resultPrefix))

	err = <-errCh
	if err != nil {
		return fmt.Errorf("error while running Factorio: %w", err)
	}

	// Remove temporary directory.
	if err := os.RemoveAll(tmpdir); err != nil {
		return fmt.Errorf("unable to remove temp dir %q: %w", tmpdir, err)
	}
	glog.Infof("temp dir %q removed", tmpdir)

	return nil
}

var cmdRender = &cobra.Command{
	Use:   "render",
	Short: "Create a screenshot from a save.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return render(cmd.Context(), factorioSettings, renderFlags, args[0])
	},
}

var renderFlags = &RenderFlags{}

func init() {
	renderFlags.Register(cmdRender.PersistentFlags(), "")
	cmdRoot.AddCommand(cmdRender)
}
