/*
Copyright (c) 2020 Red Hat, Inc.

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

package admin

import (
	"crypto/rand"
	"math/big"
	"os"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/spf13/cobra"

	"github.com/openshift/rosa/cmd/create/idp"
	"github.com/openshift/rosa/pkg/object"
	"github.com/openshift/rosa/pkg/ocm"
	"github.com/openshift/rosa/pkg/output"
	"github.com/openshift/rosa/pkg/rosa"
)

var Cmd = &cobra.Command{
	Use:   "admin",
	Short: "Creates an admin user to login to the cluster",
	Long:  "Creates a cluster-admin user with an auto-generated password to login to the cluster",
	Example: `  # Create an admin user to login to the cluster
  rosa create admin -c mycluster -p MasterKey123`,
	Run: run,
}

var args struct {
	passwordArg string
}

func init() {
	ocm.AddClusterFlag(Cmd)
	flags := Cmd.Flags()
	flags.StringVarP(
		&args.passwordArg,
		"password",
		"p",
		"",
		"Choice of password for admin user.",
	)
	output.AddFlag(Cmd)
}

func run(cmd *cobra.Command, _ []string) {
	r := rosa.NewRuntime().WithAWS().WithOCM()
	defer r.Cleanup()

	clusterKey := r.GetClusterKey()

	cluster := r.FetchCluster()
	if cluster.State() != cmv1.ClusterStateReady {
		r.Reporter.Errorf("Cluster '%s' is not yet ready", clusterKey)
		os.Exit(1)
	}

	// Try to find an existing htpasswd identity provider and
	// check if cluster-admin user already exists
	existingHTPasswdIDP, existingUserList := idp.FindExistingHTPasswdIDP(cluster, r)
	if idp.HasClusterAdmin(existingUserList) {
		r.Reporter.Errorf("Cluster '%s' already has an admin", clusterKey)
		os.Exit(1)
	}

	// No cluster admin yet: proceed to create it.
	var password string
	var err error
	passwordArg := args.passwordArg
	if len(passwordArg) == 0 {
		r.Reporter.Debugf("Generating random password")
		password, err = generateRandomPassword(23)
		if err != nil {
			r.Reporter.Errorf("Failed to generate a random password")
			os.Exit(1)
		}
	} else {
		password = passwordArg
		r.Reporter.Debugf("Using user provided password")
	}

	// Add admin user to the cluster-admins group:
	r.Reporter.Debugf("Adding '%s' user to cluster '%s'", idp.ClusterAdminUsername, clusterKey)
	user, err := cmv1.NewUser().ID(idp.ClusterAdminUsername).Build()
	if err != nil {
		r.Reporter.Errorf("Failed to create user '%s' for cluster '%s'", idp.ClusterAdminUsername, clusterKey)
		os.Exit(1)
	}

	_, err = r.OCMClient.CreateUser(cluster.ID(), "cluster-admins", user)
	if err != nil {
		r.Reporter.Errorf("Failed to add user '%s' to cluster '%s': %s",
			idp.ClusterAdminUsername, clusterKey, err)
		os.Exit(1)
	}

	// No HTPasswd IDP - create it with cluster-admin user.
	if existingHTPasswdIDP == nil {
		r.Reporter.Debugf("Adding '%s' idp to cluster '%s'", idp.HTPasswdIDPName, clusterKey)
		htpasswdIDP := cmv1.NewHTPasswdIdentityProvider().Users(cmv1.NewHTPasswdUserList().Items(
			idp.CreateHTPasswdUser(idp.ClusterAdminUsername, password),
		))
		newIDP, err := cmv1.NewIdentityProvider().
			Type(cmv1.IdentityProviderTypeHtpasswd).
			Name(idp.HTPasswdIDPName).
			Htpasswd(htpasswdIDP).
			Build()
		if err != nil {
			r.Reporter.Errorf(
				"Failed to create '%s' identity provider for cluster '%s'",
				idp.HTPasswdIDPName,
				clusterKey,
			)
			os.Exit(1)
		}

		// Add HTPasswd IDP to cluster:
		_, err = r.OCMClient.CreateIdentityProvider(cluster.ID(), newIDP)
		if err != nil {
			//since we could not add the HTPasswd IDP to the cluster, roll back and remove the cluster admin
			r.Reporter.Errorf("Failed to add '%s' identity provider to cluster '%s' as part of admin flow. "+
				"Please try again: %s", idp.HTPasswdIDPName, clusterKey, err)

			err = r.OCMClient.DeleteUser(cluster.ID(), "cluster-admins", user.ID())
			if err != nil {
				r.Reporter.Errorf("Failed to revert the admin user for cluster '%s'. Please try again: %s",
					clusterKey, err)
			}
			os.Exit(1)
		}
	} else {
		// HTPasswd IDP exists - add new cluster-admin user to it.
		r.Reporter.Debugf("Cluster has an HTPasswd IDP, will add cluster-admin to it")
		err = r.OCMClient.AddHTPasswdUser(idp.ClusterAdminUsername, password, cluster.ID(), existingHTPasswdIDP.ID())
		if err != nil {
			r.Reporter.Errorf("Failed to add user '%s' to the HTPasswd IDP of cluster '%s': %s",
				idp.ClusterAdminUsername, clusterKey, err)

			//since we could not add the cluster-admin user ot the HTPasswd IDP, roll back and remove the cluster admin
			err = r.OCMClient.DeleteUser(cluster.ID(), "cluster-admins", user.ID())
			if err != nil {
				r.Reporter.Errorf("Failed to remove user '%s' to cluster '%s': %s",
					idp.ClusterAdminUsername, clusterKey, err)
			}

			os.Exit(1)
		}
	}

	outputObject := object.Object{
		"api_url":  cluster.API().URL(),
		"username": idp.ClusterAdminUsername,
		"password": password,
	}

	if output.HasFlag() {
		if len(passwordArg) != 0 {
			delete(outputObject, "password")
		}
		err = output.Print(outputObject)
		if err != nil {
			r.Reporter.Errorf("%s", err)
			os.Exit(1)
		}
		return
	}

	r.Reporter.Infof("Admin account has been added to cluster '%s'.", clusterKey)
	r.Reporter.Infof("Please securely store this generated password. " +
		"If you lose this password you can delete and recreate the cluster admin user.")
	r.Reporter.Infof("To login, run the following command:\n\n"+
		"   oc login %s --username %s --password %s\n",
		outputObject["api_url"], outputObject["username"], outputObject["password"])
	r.Reporter.Infof("It may take several minutes for this access to become active.")
}

func generateRandomPassword(length int) (string, error) {
	const (
		lowerLetters = "abcdefghijkmnopqrstuvwxyz"
		upperLetters = "ABCDEFGHIJKLMNPQRSTUVWXYZ"
		digits       = "23456789"
		all          = lowerLetters + upperLetters + digits
	)
	var password string
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
		if err != nil {
			return "", err
		}
		newchar := string(all[n.Int64()])
		if password == "" {
			password = newchar
		}
		if i < length-1 {
			n, err = rand.Int(rand.Reader, big.NewInt(int64(len(password)+1)))
			if err != nil {
				return "", err
			}
			j := n.Int64()
			password = password[0:j] + newchar + password[j:]
		}
	}

	pw := []rune(password)
	for _, replace := range []int{5, 11, 17} {
		pw[replace] = '-'
	}

	return string(pw), nil
}
