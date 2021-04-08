package helpers

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type ClusterSetMapper struct {
	mutex sync.RWMutex
	// mapping: ClusterSet - Objects
	clusterSetToObjects map[string]sets.String
}

func NewClusterSetMapper() *ClusterSetMapper {
	return &ClusterSetMapper{
		clusterSetToObjects: make(map[string]sets.String),
	}
}

func (c *ClusterSetMapper) UpdateClusterSetByObjects(clusterSetName string, objects sets.String) {
	if clusterSetName == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.clusterSetToObjects[clusterSetName] = objects

	return
}

func (c *ClusterSetMapper) DeleteClusterSet(clusterSetName string) {
	if clusterSetName == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.clusterSetToObjects, clusterSetName)

	return
}

//DeleteObjectInClusterSet will delete cluster in all clusterset mapping
func (c *ClusterSetMapper) DeleteObjectInClusterSet(objectName string) {
	if objectName == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for clusterset, objects := range c.clusterSetToObjects {
		if !objects.Has(objectName) {
			continue
		}
		objects.Delete(objectName)
		if len(objects) == 0 {
			delete(c.clusterSetToObjects, clusterset)
		}
	}

	return
}

//UpdateObjectInClusterSet updates clusterset to cluster mapping.
//If a the clusterset of a object is changed, this func remove object from the previous mapping and add in new one.
func (c *ClusterSetMapper) UpdateObjectInClusterSet(objectName, clusterSetName string) {
	if objectName == "" || clusterSetName == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clusterSetToObjects[clusterSetName]; !ok {
		cluster := sets.NewString(objectName)
		c.clusterSetToObjects[clusterSetName] = cluster
	} else {
		c.clusterSetToObjects[clusterSetName].Insert(objectName)
	}

	for clusterset, objects := range c.clusterSetToObjects {
		if clusterSetName == clusterset {
			continue
		}
		if !objects.Has(objectName) {
			continue
		}
		objects.Delete(objectName)
		if len(objects) == 0 {
			delete(c.clusterSetToObjects, clusterset)
		}
	}

	return
}

func (c *ClusterSetMapper) GetObjectsOfClusterSet(clusterSetName string) sets.String {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.clusterSetToObjects[clusterSetName]
}

func (c *ClusterSetMapper) GetAllClusterSetToObjects() map[string]sets.String {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.clusterSetToObjects
}
