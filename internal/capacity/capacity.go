/*
Copyright © 2021 Alex Krzos akrzos@redhat.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package capacity

import "k8s.io/apimachinery/pkg/api/resource"

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ReadableCPU(cpu resource.Quantity) float64 {
	// Convert millicores to cores
	return float64(cpu.MilliValue()) / 1000
}

func ReadableMem(mem resource.Quantity) float64 {
	// Convert from KiB to GiB (Gibibyte)
	return float64(mem.Value()) / 1024 / 1024 / 1024
}

func ReadableStorage(storage resource.Quantity) float64 {
	// Convert from KiB to GB (Gigabyte)
	return float64(storage.Value()) / 1000 / 1000 / 1000
}
