//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcher) DeepCopyInto(out *CertWatcher) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcher.
func (in *CertWatcher) DeepCopy() *CertWatcher {
	if in == nil {
		return nil
	}
	out := new(CertWatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CertWatcher) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherAction) DeepCopyInto(out *CertWatcherAction) {
	*out = *in
	out.Echo = in.Echo
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherAction.
func (in *CertWatcherAction) DeepCopy() *CertWatcherAction {
	if in == nil {
		return nil
	}
	out := new(CertWatcherAction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherActionEcho) DeepCopyInto(out *CertWatcherActionEcho) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherActionEcho.
func (in *CertWatcherActionEcho) DeepCopy() *CertWatcherActionEcho {
	if in == nil {
		return nil
	}
	out := new(CertWatcherActionEcho)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherList) DeepCopyInto(out *CertWatcherList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CertWatcher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherList.
func (in *CertWatcherList) DeepCopy() *CertWatcherList {
	if in == nil {
		return nil
	}
	out := new(CertWatcherList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CertWatcherList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherSecret) DeepCopyInto(out *CertWatcherSecret) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherSecret.
func (in *CertWatcherSecret) DeepCopy() *CertWatcherSecret {
	if in == nil {
		return nil
	}
	out := new(CertWatcherSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherSpec) DeepCopyInto(out *CertWatcherSpec) {
	*out = *in
	out.Secret = in.Secret
	out.Actions = in.Actions
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherSpec.
func (in *CertWatcherSpec) DeepCopy() *CertWatcherSpec {
	if in == nil {
		return nil
	}
	out := new(CertWatcherSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertWatcherStatus) DeepCopyInto(out *CertWatcherStatus) {
	*out = *in
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertWatcherStatus.
func (in *CertWatcherStatus) DeepCopy() *CertWatcherStatus {
	if in == nil {
		return nil
	}
	out := new(CertWatcherStatus)
	in.DeepCopyInto(out)
	return out
}
