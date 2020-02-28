package filter

import (
	"github.com/goharbor/harbor/src/p2ppreheat/models"
	"github.com/goharbor/harbor/src/replication/source"
	"github.com/goharbor/harbor/src/replication"
	filter_models "github.com/goharbor/harbor/src/replication/models"
)

// Manager ...
type Manager interface {
	DoFilter(policy *models.P2PPreheatPolicy, filterItems []filter_models.FilterItem) ([]filter_models.FilterItem, error)
}

type FilterManager struct {
	// Handle the things related with source
	sourcer *source.Sourcer
}

// NewFilterManager is the constructor of FilterManager.
func NewFilterManager() *FilterManager {
	return  &FilterManager{
		sourcer:        source.NewSourcer(),
	}
}

// DoFilter is used to filter the trigger candidates by filter rules.
func (fm *FilterManager) DoFilter(policy *models.P2PPreheatPolicy, filterItems []filter_models.FilterItem) ([]filter_models.FilterItem, error) {
	// do filter
	filterChain := buildFilterChain(policy, fm.sourcer)
	return filterChain.DoFilter(filterItems), nil
}

func buildFilterChain(policy *models.P2PPreheatPolicy, sourcer *source.Sourcer) source.FilterChain {
	filters := []source.Filter{}

	fm := map[string][]filter_models.Filter{}
	for _, filter := range policy.Filters {
		fm[filter.Kind] = append(fm[filter.Kind], filter)
	}

	registry := sourcer.GetAdaptor(replication.AdaptorKindHarbor)
	// repository filter
	pattern := ""
	repoFilters := fm[replication.FilterItemKindRepository]
	if len(repoFilters) > 0 {
		pattern = repoFilters[0].Value.(string)
	}
	filters = append(filters,
		source.NewRepositoryFilter(pattern, registry))
	// tag filter
	pattern = ""
	tagFilters := fm[replication.FilterItemKindTag]
	if len(tagFilters) > 0 {
		pattern = tagFilters[0].Value.(string)
	}
	filters = append(filters,
		source.NewTagFilter(pattern, registry))
	// label filters
	var labelID int64
	for _, labelFilter := range fm[replication.FilterItemKindLabel] {
		labelID = labelFilter.Value.(int64)
		filters = append(filters, source.NewLabelFilter(labelID))
	}
	return source.NewDefaultFilterChain(filters)
}

