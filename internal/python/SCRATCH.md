

// func findSystemPythonInstallations() ([]string, error) {
//     path := os.Getenv("PATH")
//     if path == "" {
//         return nil, fmt.Errorf("PATH environment variable not set")
//     }
//     paths := strings.Split(path, string(os.PathListSeparator))
//     slog.Info("PATH", "paths", paths)
//     acc := make([]string, 0)
//     for _, dir := range paths {
//         entries, err := os.ReadDir(dir)
//         if err != nil {
//             slog.Warn("Failed to read directory in PATH", "dir", dir, "error", err)
//             continue
//         }
//         for _, entry := range entries {
//             if entry.Type().IsRegular() && strings.HasPrefix(entry.Name(), "python3") {
//                 fullPath := filepath.Join(dir, entry.Name())
//                 acc = append(acc, fullPath)
//             }
//         }
//     }
//     return acc, nil
// }
	
// Find all system python installations
// systemPythons, err := findSystemPythonInstallations()
// if err != nil {
//     slog.Error("Error finding system Python installations", "error", err)
//     return err
// }
// for _, python := range systemPythons {
//     slog.Info("Found system Python installation", "python", python)
//     sitePackages, err := findPythonSitePackages(python)
//     if err != nil {
//         slog.Error("Error finding Python site-packages", "python", python, "error", err)
//         return err
//     }
//     slog.Info("Found Python site-packages", "python", python, "sitePackages", sitePackages)
// }

// type Library struct {
// 	Name            string
// 	Version         string
// 	Path            string
// 	TopLevelModules []string
// }
//
//
// func topLevelModulesFromDistInfo(path string) ([]string, error) {
// 	acc := make([]string, 0)
// 	topLevelFile := filepath.Join(path, "top_level.txt")
// 	if _, err := os.Stat(topLevelFile); err == nil {
// 		// Read the top_level.txt file to get the library name
// 		data, err := os.ReadFile(topLevelFile)
// 		if err != nil {
// 			slog.Warn("Error reading top_level.txt file", "path", topLevelFile, "error", err)
// 			return nil, err
// 		}
// 		// N.b. there may be multiple top level modules, so we need to split them
// 		lines := strings.SplitSeq(string(data), "\n")
// 		for line := range lines {
// 			line = strings.TrimSpace(line)
// 			if line != "" && !strings.HasPrefix(line, "_") {
// 				acc = append(acc, line)
// 			}
// 		}
// 	} else {
// 		// If that doesn't exist, use the directory name until the last hyphen
// 		name := strings.Replace(filepath.Base(path), ".dist-info", "", 1)
// 		name = name[:strings.LastIndex(name, "-")]
// 		if !strings.HasPrefix(name, "_") {
// 			acc = append(acc, name)
// 		}
// 	}
// 	return acc, nil
// }
//
// // TODO This might be insufficient. The best way is probably to read the __version__ property of a
// // module, but that would require loading it up via python.
// func libraryVersionFromDistInfo(path string) string {
// 	name := filepath.Base(path)
// 	name = strings.Replace(name, ".dist-info", "", 1)
// 	version := name[strings.LastIndex(name, "-")+1:]
// 	return version
// }
//
// func libraryNameFromDistInfo(path string) string {
// 	name := filepath.Base(path)
// 	name = strings.Replace(name, ".dist-info", "", 1)
// 	name = name[:strings.LastIndex(name, "-")]
// 	return name
// }

// func findLibraries(sitePackagesDir string) ([]Library, error) {
// 	acc := make([]Library, 0)
// 	err := filepath.WalkDir(sitePackagesDir, func(path string, d fs.DirEntry, err error) error {
// 		if d.IsDir() && strings.HasSuffix(d.Name(), ".dist-info") {
// 			name := libraryNameFromDistInfo(path)
// 			version := libraryVersionFromDistInfo(path)
// 			topLevelModules, err := topLevelModulesFromDistInfo(path)
// 			if err != nil {
// 				slog.Warn("Error getting library name", "path", path, "error", err)
// 				return err
// 			}
// 			acc = append(acc, Library{
// 				Name:            name,
// 				Version:         version,
// 				Path:            path,
// 				TopLevelModules: topLevelModules,
// 			})
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		slog.Warn("Error finding Python libraries", "error", err)
// 		return nil, err
// 	}
// 	return acc, nil
// }

// func findModules(library Library) ([]Module, error) {
// 	acc := make([]Module, 0)
// 	// Get the parent directory of the dist-info directory
// 	sitePackagesDir := filepath.Dir(library.Path)
// 	// Look for all of the top level modules in that directory
// 	for _, topLevelModule := range library.TopLevelModules {
// 		moduleDir := filepath.Join(sitePackagesDir, topLevelModule)
// 		if _, err := os.Stat(moduleDir); err != nil {
// 			slog.Warn("Error checking module directory", "path", moduleDir, "error", err)
// 			continue
// 		}
// 		err := filepath.WalkDir(moduleDir, func(path string, d fs.DirEntry, err error) error {
// 			// Find all the __init__.py files in the library directory
// 			if err != nil {
// 				slog.Warn("Error walking directory", "path", path, "error", err)
// 				return err
// 			}
// 			if !d.IsDir() && d.Name() == "__init__.py" {
// 				init := strings.Replace(path, sitePackagesDir, "", 1)
// 				initParts := strings.Split(init, string(os.PathSeparator))
// 				initParts = initParts[1 : len(initParts)-1]
// 				name := strings.Join(initParts, ".")
// 				acc = append(acc, Module{
// 					// Library: library,
// 					Name: name,
// 					Path: path,
// 				})
// 			}
// 			return nil
// 		})
// 		if err != nil {
// 			slog.Warn("Error finding Python modules", "error", err)
// 			return nil, err
// 		}
// 	}
// 	return acc, nil
// }

// libraries, err := findLibraries(sitePackagesDir)
// if err != nil {
// 	slog.Error("Error finding Python libraries", "sitePackagesDir", sitePackagesDir, "error", err)
// 	return err
// }
// for _, lib := range libraries {
// 	slog.Debug("Found Python library", "lib", lib)
//           modules, err := findModules(lib)
//           if err != nil {
//               slog.Error("Error finding Python modules", "lib", lib.Name, "error", err)
//               return err
//           }
//           for _, module := range modules {
//               slog.Info("Found Python module", "module", module)
//           }
// }

// func findSitePackagesDir(root string) (string, error) {
// 	// Find site-packages directory
// 	var sitePackagesDir string
// 	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			slog.Warn("Error walking directory", "path", path, "error", err)
// 			return nil
// 		}
// 		if d.IsDir() && d.Name() == "site-packages" {
// 			sitePackagesDir = path
// 			return fs.SkipDir
// 		}
// 		return nil
// 	})
// 	// Check if site-packages directory was found
// 	if err != nil || sitePackagesDir == "" {
// 		slog.Warn("Error finding site-packages directory", "error", err)
// 		return "", err
// 	}
// 	// Find libraries in site-packages directory
// 	return sitePackagesDir, nil
// }
//
// func findSitePackages(sitePackagesDir string) error {
//     // Look for all directories in site-packages that have an __init__.py file
//     moduleInitFiles := make([]string, 0)
//     files, err := os.ReadDir(sitePackagesDir)
//     if err != nil {
//         slog.Error("Error reading directory", "path", sitePackagesDir, "error", err)
//         return err
//     }
//     for _, file := range files {
//         initFile := filepath.Join(sitePackagesDir, file.Name(), "__init__.py")
//         if _, err := os.Stat(initFile); err == nil {
//             moduleInitFiles = append(moduleInitFiles, initFile)
//         }
//     }
//     for _, initFile := range moduleInitFiles {
//         slog.Info("Found init file", "initFile", initFile)
//     }
//     // Find the version of the package
//     // Find all the modules underneath these packages
//     return nil
// }
//

// func findVirtualEnvironmentModules(sitePackagesDir string) ([]Module, error) {
//     // Find init files
// 	initFiles := make([]string, 0)
// 	filepath.WalkDir(sitePackagesDir, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !d.IsDir() && d.Name() == "__init__.py" {
//             initFiles = append(initFiles, path)
// 		}
// 		return nil
// 	})
//     // Use init files to identify all module files
//     slog.Info("Found init files", "initFiles", initFiles)
//     acc := make([]Module, 0)
//     for _, initFile := range initFiles {
//         dir := filepath.Dir(initFile)
//         files, err := os.ReadDir(dir)
//         if err != nil {
//             slog.Error("Error reading directory", "path", filepath.Dir(initFile), "error", err)
//             return nil, err
//         }
//         for _, file := range files {
//             if strings.HasSuffix(file.Name(), ".py") {
//                 path := filepath.Join(dir, file.Name())
//                 path = strings.Replace(path, sitePackagesDir, "", 1)
//                 path = strings.TrimPrefix(path, string(os.PathSeparator))
//                 path = strings.TrimSuffix(path, "/__init__.py")
//                 path = strings.TrimSuffix(path, ".py")
//                 name := strings.ReplaceAll(path, string(os.PathSeparator), ".")
//                 acc = append(acc, Module{
//                     Name: name,
//                     Path: filepath.Join(dir, file.Name()),
//                 })
//             }
//         }
//     }
// 	return acc, nil
// }

    // sitePackagesDir, err := findSitePackagesDir(venv)
    // if err != nil {
    // 	slog.Error("Error finding Python site-packages", "venv", venv, "error", err)
    // 	return err
    // }
//       findSitePackages(sitePackagesDir)
    // modules, err := findVirtualEnvironmentModules(sitePackagesDir)
    // if err != nil {
    // 	slog.Error("Error finding Python modules", "sitePackagesDir", sitePackagesDir, "error", err)
    // 	return err
    // }
    // for _, module := range modules {
    // 	slog.Info("Found Python module", "module", module)
    // }

type Module struct {
	Name string
	Path string
}

func pipFreeze(venv string) ([]Dependency, error) {
	acc := make([]Dependency, 0)
	binary := filepath.Join(venv, "bin", "python")
	cmd := exec.Command(binary, "-m", "pip", "freeze")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Error running pip freeze", "venv", venv, "error", err)
		return nil, err
	}
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		nameVersion := strings.Split(line, "==")
		if len(nameVersion) != 2 {
			slog.Warn("Invalid line format", "line", line)
			continue
		}
		name := strings.TrimSpace(nameVersion[0])
		version := strings.TrimSpace(nameVersion[1])
		acc = append(acc, Dependency{
			Name:    name,
			Version: version,
		})
	}
	return acc, nil
}

			// distInfo, err := findPackageDistInfo(sitePackagesDir, pkg)
			// if err != nil {
			//     slog.Error("Error finding package dist-info", "package", pkg, "error", err)
			//     continue
			// }
			// slog.Info("Found dist-info", "distInfo", distInfo)
		// Dependencies
		// dependencies, err := pipFreeze(venv)
		// if err != nil {
		// 	slog.Error("Error running pip freeze", "venv", venv, "error", err)
		// 	continue
		// }
		//       // TODO Find the METADATA modules
		// for _, dep := range dependencies {
		// 	slog.Info("GB_TEST", "dep", dep)
		// }


		// sitePackagesDir, err := findSitePackagesDir(venv)
		// if err != nil {
		// 	slog.Error("Error finding sitePackagesDir", "sitePackagesDir", sitePackagesDir)
		// 	continue
		// }

		// packages, err := findSitePackageMetadata(sitePackagesDir)
		// if err != nil {
		// 	slog.Error("Error finding site package metadata",
		// 		"sitePackagesDir", sitePackagesDir, "error", err)
		// 	continue
		// }
		// for _, pkg := range packages {
		//           slog.Info("Found package", "name", pkg.Name)
		//           for _, module := range pkg.Modules {
		//               slog.Info("\tModule:", "module", module)
		//           }
		// }

// func moduleFromTopLevelFile(path string) ([]string, error) {
// 	rootModules := make([]string, 0)
// 	if common.Exists(path) {
// 		topLevel, err := os.ReadFile(path)
// 		if err != nil {
// 			slog.Error("Error opening top_level.txt file", "topLevel", topLevel, "error", err)
// 			return nil, err
// 		}
// 		for line := range strings.SplitSeq(string(topLevel), "\n") {
// 			line = strings.TrimSpace(line)
// 			if line == "" {
// 				continue
// 			}
// 			rootModules = append(rootModules, line)
// 		}
// 	}
// 	return rootModules, nil
// }

// func findInitFiles(sitePackagesDir string, rootModules []string) ([]string, error) {
// 	initFiles := make([]string, 0)
// 	for _, rootModule := range rootModules {
// 		rootModulePath := filepath.Join(sitePackagesDir, rootModule)
// 		if !common.Exists(rootModulePath) {
// 			slog.Debug("Root module path does not exist", "rootModulePath", rootModulePath)
// 			continue
// 		}
// 		err := filepath.WalkDir(
// 			rootModulePath,
// 			func(path string, d fs.DirEntry, err error) error {
// 				if err != nil {
// 					return err
// 				}
// 				if !d.IsDir() && d.Name() == "__init__.py" {
// 					initFiles = append(initFiles, path)
// 				}
// 				return nil
// 			})
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return initFiles, nil
// }

// func findPythonFiles(initFiles []string) ([]string, error) {
// 	allFiles := make([]string, 0)
// 	for _, initFile := range initFiles {
// 		dir := filepath.Dir(initFile)
// 		files, err := os.ReadDir(dir)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, file := range files {
// 			if strings.HasSuffix(file.Name(), ".py") {
// 				allFiles = append(allFiles, filepath.Join(dir, file.Name()))
// 			}
// 		}
// 	}
// 	return allFiles, nil
// }

// func parseMetadataFile(sitePackagesDir string, path string) (*Package, error) {
// 	// Read the METADATA file
// 	file, err := os.Open(path)
// 	if err != nil {
// 		slog.Error("Error opening METADATA file", "path", path, "error", err)
// 		return nil, err
// 	}
// 	defer file.Close()
// 	var name, version string
// 	// Read Name and Version from METADATA file
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.HasPrefix(line, "Name: ") {
// 			name = strings.TrimPrefix(line, "Name: ")
// 		} else if strings.HasPrefix(line, "Version: ") {
// 			version = strings.TrimPrefix(line, "Version: ")
// 		}
// 	}
// 	if name == "" || version == "" {
// 		slog.Warn("Invalid METADATA file", "path", path)
// 		return nil, fmt.Errorf("invalid METADATA file")
// 	}
// 	// Find rootModules
// 	rootModules := make([]string, 0)
// 	// Read the top_level.txt file
// 	distInfo := filepath.Dir(path)
// 	topLevelFile := filepath.Join(distInfo, "top_level.txt")
// 	topLevelModules, err := moduleFromTopLevelFile(topLevelFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	rootModules = append(rootModules, topLevelModules...)
// 	// Check if the name contains the module
// 	distInfoName := filepath.Base(distInfo)
// 	distInfoName = strings.TrimSuffix(distInfoName, ".dist-info")
// 	distInfoName = distInfoName[:strings.LastIndex(distInfoName, "-")]
// 	if common.AtLeastOneFileExists([]string{
// 		filepath.Join(sitePackagesDir, fmt.Sprintf("%s.py", distInfoName)), // It is a single python file
// 		filepath.Join(sitePackagesDir, distInfoName, "__init__.py"),        // It is a directory module
// 	}) {
// 		rootModules = append(rootModules, distInfoName)
// 	}
// 	// TODO Does this actually get what we want?
// 	// Check RECORD file
// 	// recordFileName := filepath.Join(distInfo, "RECORD")
// 	//    if common.Exists(recordFileName) {
// 	// 	// Open RECORD file
// 	// 	recordFile, recordFileErr := os.Open(recordFileName)
// 	// 	if recordFileErr != nil {
// 	// 		return nil, recordFileErr
// 	// 	}
// 	// 	defer recordFile.Close()
// 	// 	// Scan through and find unique first path elements (i.e. root modules)
// 	// 	acc := make(map[string]bool, 0)
// 	// 	scanner = bufio.NewScanner(recordFile)
// 	// 	for scanner.Scan() {
// 	// 		line := scanner.Text()
// 	// 		parts := strings.Split(line, string(os.PathSeparator))
// 	// 		if len(parts) > 0 {
// 	// 			acc[parts[0]] = true
// 	// 		}
// 	// 	}
// 	// 	// Add these to the root modules
// 	// 	for rootModule := range acc {
// 	// 		if !strings.HasSuffix(rootModule, ".dist-info") &&
// 	// 			rootModule != "__pycache__" &&
// 	// 			filepath.IsAbs(rootModule) {
// 	// 			rootModules = append(rootModules, rootModule)
// 	// 		}
// 	// 	}
// 	// }
// 	// Handle irregular cases
// 	irregularCases := map[string]string{
// 		"beautifulsoup4":           "bs4",
// 		"protobuf":                 "proto",
// 		"PyYAML":                   "yaml",
// 		"python-dateutil":          "dateutil",
// 		"scikit-learn":             "sklearn",
// 		"Pillow":                   "PIL",
// 		"backports.zoneinfo":       "backports.zoneinfo",
// 		"google-api-python-client": "googleapiclient",
// 		"python-decouple":          "decouple",
// 		"dnspython":                "dns",
// 	}
// 	if irregularCases[name] != "" {
// 		rootModules = append(rootModules, irregularCases[name])
// 	}
// 	// Find all of the __init__.py files in the root modules
// 	initFiles, err := findInitFiles(sitePackagesDir, rootModules)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Find all the other files in those directories
// 	allFiles, err := findPythonFiles(initFiles)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Find any single file modules
// 	singleFileModule := filepath.Join(sitePackagesDir, fmt.Sprintf("%s.py", distInfoName))
// 	if common.Exists(singleFileModule) {
// 		allFiles = append(allFiles, singleFileModule)
// 	}
// 	// Transform all the files into modules
// 	allModules := make([]string, 0)
// 	for _, file := range allFiles {
// 		allModules = append(allModules, moduleFromPath(strings.ReplaceAll(file, sitePackagesDir, "")))
// 	}
// 	return &Package{
// 		Name:        name,
// 		Version:     version,
// 		DistInfo:    distInfo,
// 		RootModules: common.Dedupe(rootModules),
// 		Modules:     common.Dedupe(allModules),
// 	}, nil
// }

// func findSitePackageMetadata(sitePackagesDir string) ([]*Package, error) {
// 	acc := make([]*Package, 0)
// 	err := filepath.WalkDir(sitePackagesDir, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			slog.Warn("Error walking directory", "path", path, "error", err)
// 			return nil
// 		}
// 		if !d.IsDir() && d.Name() == "METADATA" {
// 			metadata, err := parseMetadataFile(sitePackagesDir, path)
// 			if err != nil {
// 				slog.Error("Error parsing METADATA file", "path", path, "error", err)
// 				return nil
// 			}
// 			acc = append(acc, metadata)
// 			return fs.SkipDir
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		slog.Warn("Error finding METADATA files", "error", err)
// 		return nil, err
// 	}
// 	return acc, nil
// }


// type Package struct {
// 	Name        string
// 	Version     string
// 	DistInfo    string
// 	RootModules []string
// 	Modules     []string
// }

// type Dependency struct {
// 	Name    string
// 	Version string
// }
