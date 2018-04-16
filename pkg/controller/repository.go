package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/appscode/stash/apis/stash/v1alpha1"
	stash_util "github.com/appscode/stash/client/clientset/versioned/typed/stash/v1alpha1/util"
	"github.com/appscode/stash/pkg/osm"
	"github.com/appscode/stash/pkg/util"
	"github.com/golang/glog"
	"github.com/graymeta/stow"
)

func (c *StashController) initRepositoryWatcher() {
	c.repoInformer = c.stashInformerFactory.Stash().V1alpha1().Repositories().Informer()
	c.repoQueue = queue.New("Repository", c.MaxNumRequeues, c.NumThreads, c.runRepositoryInjector)
	c.repoInformer.AddEventHandler(queue.DefaultEventHandler(c.repoQueue.GetQueue()))
	c.repoLister = c.stashInformerFactory.Stash().V1alpha1().Repositories().Lister()
}

func (c *StashController) runRepositoryInjector(key string) error {
	obj, exist, err := c.repoInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("Repository %s does not exist anymore\n", key)
	} else {
		glog.Infof("Sync/Add/Update for Repository %s\n", key)

		repo := obj.(*api.Repository)

		if repo.DeletionTimestamp != nil {
			if core_util.HasFinalizer(repo.ObjectMeta, util.RepositoryFinalizer) {
				err = c.deleteResticRepository(repo)
				if err != nil {
					return err
				}
				_, _, err = stash_util.PatchRepository(c.stashClient.StashV1alpha1(), repo, func(in *api.Repository) *api.Repository {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, util.RepositoryFinalizer)
					return in
				})
				if err != nil {
					return err
				}
			}
		} else {
			if repo.Spec.WipeOut {
				_, _, err = stash_util.PatchRepository(c.stashClient.StashV1alpha1(), repo, func(in *api.Repository) *api.Repository {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, util.RepositoryFinalizer)
					return in
				})
			}
		}
	}
	return nil
}

func (c *StashController) deleteResticRepository(repository *api.Repository) error {
	cfg,err:=osm.NewOSMContext(c.kubeClient,repository)
	if err!=nil{
		return err
	}

	loc,err :=stow.Dial(cfg.Provider,cfg.Config)
	if err!=nil{
		return err
	}


	bucket,err:= util.GetBucket(&repository.Spec.Backend)
	if err!=nil{
		return err
	}

	container,err:=loc.Container(bucket)
	if err!=nil{
		return err
	}


	return nil
}
