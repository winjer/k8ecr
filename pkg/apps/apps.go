package apps

import (
	"k8s.io/client-go/kubernetes"
)

// App is a group of images mapped to containers within resources in the app
type App struct {
	Name       string
	ChangeSets map[ImageIdentifier]ChangeSet
}

// NewApp returns a new App
func NewApp(name string) App {
	return App{
		Name:       name,
		ChangeSets: make(map[ImageIdentifier]ChangeSet),
	}
}

// GetChangeSets returns all changesets in the App
func (app *App) GetChangeSets() []ChangeSet {
	images := make([]ChangeSet, 0)
	for _, v := range app.ChangeSets {
		images = append(images, v)
	}
	return images
}

// SetLatest sets the latest version on every changeset in this app
func (app *App) SetLatest(registry, repository, version string) {
	id := ImageIdentifier{Registry: registry, Repo: repository}
	cs, ok := app.ChangeSets[id]
	if ok {
		cs.SetLatest(version)
	}

}

// AppManager finds and updates Applications
// and their deployments and cronjobs
type AppManager struct {
	ClientSet kubernetes.Interface
	Namespace string
	Apps      map[string]App
	Managers  map[string]*ResourceManager
}

// NewAppManager creates a new Image manager
func NewAppManager(namespace string) (*AppManager, error) {
	clientset, err := getClientSet()
	if err != nil {
		return nil, err
	}
	a := &AppManager{
		ClientSet: clientset,
		Namespace: namespace,
		Apps:      make(map[string]App),
		Managers:  resourceManagers,
	}
	err = a.Scan()
	return a, err
}

// SetLatest calls SetLatest on all contained apps
// Setting the version on all relevant containers
func (mgr *AppManager) SetLatest(registry, repository, version string) {
	for _, app := range mgr.Apps {
		app.SetLatest(registry, repository, version)
	}
}

func groupResources(resources map[string][]Container) map[string]App {
	apps := make(map[string]App)
	for kind, resources := range resources {
		for _, item := range resources {
			app, ok := apps[item.App]
			if !ok {
				app = NewApp(item.App)
			}
			image, ok := app.ChangeSets[item.ImageID]
			if !ok {
				image = NewChangeSet(item.ImageID)
			}
			image.Containers[kind] = append(image.Containers[kind], item)
			app.ChangeSets[item.ImageID] = image
			apps[app.Name] = app
		}
	}
	return apps
}

// Scan the cluster and find all resources and containers we manage
func (mgr *AppManager) Scan() error {
	resources := make(map[string][]Container)
	for _, rm := range resourceManagers {
		resources[rm.Kind] = make([]Container, 0)
		items, err := rm.Fetcher(mgr)
		if err != nil {
			return err
		}
		for _, item := range items {
			for _, r := range rm.Generator(item) {
				resources[rm.Kind] = append(resources[rm.Kind], r)
			}
		}
	}
	mgr.Apps = groupResources(resources)
	return nil
}