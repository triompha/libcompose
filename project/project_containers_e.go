package project

import (
	"fmt"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/triompha/libcompose/project/events"
)

// Containers lists the containers for the specified services. Can be filter using
// the Filter struct.
func (p *Project) ContainersE(ctx context.Context, filter Filter, services ...string) ([]ContainerShort, error) {
	containers := []ContainerShort{}
	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.NoEvent, events.NoEvent, func(service Service) error {
			serviceContainers, innerErr := service.Containers(ctx)
			service.Name()
			if innerErr != nil {
				return innerErr
			}

			for _, container := range serviceContainers {
				running, innerErr := container.IsRunning(ctx)
				if innerErr != nil {
					log.Error(innerErr)
				}
				switch filter.State {
				case Running:
					if !running {
						continue
					}
				case Stopped:
					if running {
						continue
					}
				case AnyState:
					// Don't do a thing
				default:
					// Invalid state filter
					return fmt.Errorf("Invalid container filter: %s", filter.State)
				}
				containerID, innerErr := container.ID()
				if innerErr != nil {
					log.Error(innerErr)
				}

				isRunning, innerErr := container.IsRunning(ctx)
				if innerErr != nil {
					log.Error(innerErr)
				}

				containers = append(containers, ContainerShort{
					ID:          containerID,
					IsRunning:   isRunning,
					Name:        container.Name(),
					ServiceName: service.Name(),
				})
			}
			return nil
		})
	}), nil)
	if err != nil {
		return nil, err
	}
	return containers, nil
}
