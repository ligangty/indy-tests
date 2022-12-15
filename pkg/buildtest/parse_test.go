/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package buildtest

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var TEST_CONTENT = `
+ mvn -Prelease -DskipTests=true -Djavadocs -Djavadocs.dist.skip -Dmanagement-model-documentation deploy -Dinsecure.repositories=WARN
Picked up JAVA_TOOL_OPTIONS: -Xmx1024m -Xss1m
[INFO] Scanning for projects...
[INFO] Downloading from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/group/build-97241/org/jboss/jboss-parent/35/jboss-parent-35.pom
[INFO] Downloaded from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/group/build-97241/org/jboss/jboss-parent/35/jboss-parent-35.pom (66 kB at 325 kB/s)
[INFO] Downloading from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/group/build-97241/org/wildfly/core/wildfly-core-parent/10.1.21.Final-redhat-00001/wildfly-core-parent-10.1.21.Final-redhat-00001.pom
[INFO] Downloaded from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/group/build-97241/org/wildfly/core/wildfly-core-parent/10.1.21.Final-redhat-00001/wildfly-core-parent-10.1.21.Final-redhat-00001.pom (118 kB at 872 kB/s)
[INFO] Activating AltDeploy extension 1.6 ( SHA: f74e8678 ) 
...
[INFO] --- maven-deploy-plugin:2.8.2:deploy (default-deploy) @ jboss-eap-parent ---
[INFO] Using alternate deployment repository indy-mvn::default::http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241
[INFO] Uploading to indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/7.3.8.GA-redhat-00001/jboss-eap-parent-7.3.8.GA-redhat-00001.pom
[INFO] Uploaded to indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/7.3.8.GA-redhat-00001/jboss-eap-parent-7.3.8.GA-redhat-00001.pom (454 kB at 3.7 MB/s)
[INFO] Downloading from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/maven-metadata.xml
[INFO] Downloaded from indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/maven-metadata.xml (400 B at 6.7 kB/s)
[INFO] Uploading to indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/maven-metadata.xml
[INFO] Uploaded to indy-mvn: http://indyhost/api/folo/track/build-97241/maven/hosted/build-97241/org/jboss/eap/jboss-eap-parent/maven-metadata.xml (384 B at 6.9 kB/s)
`

func TestParseLog(t *testing.T) {
	Convey("TestParseLog", t, func() {
		entries, err := ParseLog(TEST_CONTENT)
		Convey("Should not fail for parse", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should have download entries", func() {
			downs := entries["downloads"]
			So(len(downs), ShouldEqual, 3)
		})
		Convey("Should have upload entries", func() {
			ups := entries["uploads"]
			So(len(ups), ShouldEqual, 2)
		})
	})

}
