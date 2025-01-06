package scheduler

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/sidra-api/sidra/dto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (j *Job) getIngress() {
	config, err := rest.InClusterConfig()
	if err != nil {		
		return		
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Default().Println("Error get clientset: ", err)
	}
	ingresses, err := clientset.NetworkingV1().Ingresses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Default().Println("Error get ingresses: ", err)
	}
	
	if len(ingresses.Items) > 0 {		
		for _, ing := range ingresses.Items {
			if *ing.Spec.IngressClassName != "sidra" {
				continue
			}
			for _, rule := range ing.Spec.Rules {				
				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						key := rule.Host + path.Path
						j.dataSet.SerializeRoute[key] = dto.SerializeRoute{							
							Host:         rule.Host,
							Plugins:  	  ing.Annotations["konghq.com/plugins"],
							UpstreamHost: path.Backend.Service.Name,
							UpstreamPort: strconv.Itoa(int(path.Backend.Service.Port.Number)),
							Path:         path.Path,
							PathType:     string(*path.PathType),
						}
						fmt.Println("added route", key, j.dataSet.SerializeRoute[key])
					}
				}
			}
		}
	}
}