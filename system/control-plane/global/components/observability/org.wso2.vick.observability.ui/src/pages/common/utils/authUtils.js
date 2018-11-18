/*
 * Copyright (c) 2018, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * Common utilities.
 */
class AuthUtils {

    /**
     * Sign in the user.
     *
     * @param {string} user The user to be signed in
     */
    static signIn(user) {
        localStorage.setItem("user", user);
    }

    /**
     * Sign out the current user.
     */
    static signOut() {
        localStorage.removeItem("user");
    }

    /**
     * Get the currently authenticated user.
     *
     * @returns {string} The current user
     */
    static getAuthenticatedUser() {
        return localStorage.getItem("user");
    }

}

export default AuthUtils;
