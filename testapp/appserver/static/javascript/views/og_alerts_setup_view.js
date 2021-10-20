"use strict";

define(
    ["backbone", "jquery", "splunkjs/splunk"],
    function(Backbone, jquery, splunk_js_sdk) {
        sdk = splunk_js_sdk;
        var ExampleView = Backbone.View.extend({
            // -----------------------------------------------------------------
            // Backbon Functions, These are specific to the Backbone library
            // -----------------------------------------------------------------
            initialize: function initialize() {
                Backbone.View.prototype.initialize.apply(this, arguments);
            },

            events: {
                "click .setup_button": "trigger_setup",
            },

            render: function() {
                this.el.innerHTML = this.get_template();

                return this;
            },

            // -----------------------------------------------------------------
            // Custom Functions, These are unrelated to the Backbone functions
            // -----------------------------------------------------------------
            // ----------------------------------
            // Main Setup Logic
            // ----------------------------------
            // This performs some sanity checking and cleanup on the inputs that
            // the user has provided before kicking off main setup process
            trigger_setup: function trigger_setup() {
                // Used to hide the error/success outputs, when a setup is retried
                this.remove_outputs();

                console.log("Triggering setup");
                var api_key_input = jquery("input[name=api_key]");
                var api_key = api_key_input.val();
                var sanitized_api_key = this.sanitize_string(api_key);

                var region_input = jquery("input[name=region]:checked");
                var region = region_input.val();

                var error_messages_to_display = this.validate_inputs(sanitized_api_key, region);

                if (error_messages_to_display.length > 0) {
                    // Displays the errors that occurred input validation
                    this.display_error_output(error_messages_to_display);
                } else {
                    this.perform_setup(
                        splunk_js_sdk,
                        sanitized_api_key,
                        region
                    );
                }
            },

            // This is where the main setup process occurs
            perform_setup: async function perform_setup(splunk_js_sdk, api_key, region) {
                var app_name = "opsgenie";

                var application_name_space = {
                    owner: "nobody",
                    app: app_name,
                    sharing: "app",
                };

                try {
                    // Create the Splunk JS SDK Service object
                    splunk_js_sdk_service = this.create_splunk_js_sdk_service(
                        splunk_js_sdk,
                        application_name_space,
                    );
                    // Creates the passwords.conf stanza that is the encryption
                    // of the api_key provided by the user
                    await this.encrypt_api_key(splunk_js_sdk_service, api_key, region);

                    // Completes the setup, by access the app.conf's [install]
                    // stanza and then setting the `is_configured` to true
                    await this.complete_setup(splunk_js_sdk_service);

                    // Reloads the splunk app so that splunk is aware of the
                    // updates made to the file system
                    await this.reload_splunk_app(splunk_js_sdk_service, app_name);

                    // Show success message
                    this.display_success_output();
                } catch (error) {
                    // This could be better error catching.
                    // Usually, error output that is ONLY relevant to the user
                    // should be displayed. This will return output that the
                    // user does not understand, causing them to be confused.
                    var error_messages_to_display = [];
                    if (
                        error !== null &&
                        typeof error === "object" &&
                        error.hasOwnProperty("responseText")
                    ) {
                        var response_object = JSON.parse(error.responseText);
                        error_messages_to_display = this.extract_error_messages(
                            response_object.messages,
                        );
                    } else {
                        // Assumed to be string
                        error_messages_to_display.push(error);
                    }

                    this.display_error_output(error_messages_to_display);
                }
            },

            encrypt_api_key: async function encrypt_api_key(
                splunk_js_sdk_service,
                api_key,
                region,
            ) {
                // /servicesNS/<NAMESPACE_USERNAME>/<SPLUNK_APP_NAME>/storage/passwords/<REALM>%3A<USERNAME>%3A
                // /servicesNS/nobody/opsgenie/storage/passwords/us%3Aapi_key%3A
                var realm = region;
                var username = "api_key";

                var storage_passwords_accessor = splunk_js_sdk_service.storagePasswords(
                    {
                        // No namespace information provided
                    },
                );
                
                await storage_passwords_accessor.fetch();

                var does_us_storage_password_exist = this.does_storage_password_exist(
                    storage_passwords_accessor,
                    "us",
                    username,
                );

                if (does_us_storage_password_exist) {
                    await this.delete_storage_password(
                        storage_passwords_accessor,
                        "us",
                        username,
                    );
                }
                await storage_passwords_accessor.fetch();

                var does_eu_storage_password_exist = this.does_storage_password_exist(
                    storage_passwords_accessor,
                    "eu",
                    username,
                );

                if (does_eu_storage_password_exist) {
                    await this.delete_storage_password(
                        storage_passwords_accessor,
                        "eu",
                        username,
                    );
                }
                await storage_passwords_accessor.fetch();

                await this.create_storage_password_stanza(
                    storage_passwords_accessor,
                    realm,
                    username,
                    api_key,
                );
            },

            complete_setup: async function complete_setup(splunk_js_sdk_service) {
                var app_name = "opsgenie";
                var configuration_file_name = "app";
                var stanza_name = "install";
                var properties_to_update = {
                    is_configured: "true",
                };

                await this.update_configuration_file(
                    splunk_js_sdk_service,
                    configuration_file_name,
                    stanza_name,
                    properties_to_update,
                );
            },

            reload_splunk_app: async function reload_splunk_app(
                splunk_js_sdk_service,
                app_name,
            ) {
                var splunk_js_sdk_apps = splunk_js_sdk_service.apps();
                await splunk_js_sdk_apps.fetch();

                var current_app = splunk_js_sdk_apps.item(app_name);
                current_app.reload();
            },

            // ----------------------------------
            // Splunk JS SDK Helpers
            // ----------------------------------
            // ---------------------
            // Process Helpers
            // ---------------------
            update_configuration_file: async function update_configuration_file(
                splunk_js_sdk_service,
                configuration_file_name,
                stanza_name,
                properties,
            ) {
                // Retrieve the accessor used to get a configuration file
                var splunk_js_sdk_service_configurations = splunk_js_sdk_service.configurations(
                    {
                        // Name space information not provided
                    },
                );
                await splunk_js_sdk_service_configurations.fetch();

                // Check for the existence of the configuration file being editect
                var does_configuration_file_exist = this.does_configuration_file_exist(
                    splunk_js_sdk_service_configurations,
                    configuration_file_name,
                );

                // If the configuration file doesn't exist, create it
                if (!does_configuration_file_exist) {
                    await this.create_configuration_file(
                        splunk_js_sdk_service_configurations,
                        configuration_file_name,
                    );
                }

                // Retrieves the configuration file accessor
                var configuration_file_accessor = this.get_configuration_file(
                    splunk_js_sdk_service_configurations,
                    configuration_file_name,
                );
                await configuration_file_accessor.fetch();

                // Checks to see if the stanza where the inputs will be
                // stored exist
                var does_stanza_exist = this.does_stanza_exist(
                    configuration_file_accessor,
                    stanza_name,
                );

                // If the configuration stanza doesn't exist, create it
                if (!does_stanza_exist) {
                    await this.create_stanza(configuration_file_accessor, stanza_name);
                }
                // Need to update the information after the creation of the stanza
                await configuration_file_accessor.fetch();

                // Retrieves the configuration stanza accessor
                var configuration_stanza_accessor = this.get_configuration_file_stanza(
                    configuration_file_accessor,
                    stanza_name,
                );
                await configuration_stanza_accessor.fetch();

                // We don't care if the stanza property does or doesn't exist
                // This is because we can use the
                // configurationStanza.update() function to create and
                // change the information of a property
                await this.update_stanza_properties(
                    configuration_stanza_accessor,
                    properties,
                );
            },

            // ---------------------
            // Existence Functions
            // ---------------------
            does_configuration_file_exist: function does_configuration_file_exist(
                configurations_accessor,
                configuration_file_name,
            ) {
                var was_configuration_file_found = false;

                var configuration_files_found = configurations_accessor.list();
                for (var index = 0; index < configuration_files_found.length; index++) {
                    var configuration_file_name_found =
                        configuration_files_found[index].name;
                    if (configuration_file_name_found === configuration_file_name) {
                        was_configuration_file_found = true;
                    }
                }

                return was_configuration_file_found;
            },

            does_stanza_exist: function does_stanza_exist(
                configuration_file_accessor,
                stanza_name,
            ) {
                var was_stanza_found = false;

                var stanzas = configuration_file_accessor.list();
                for (var index = 0; index < stanzas.length; index++) {
                    if (stanzas[index].name === stanza_name) {
                        was_stanza_found = true;
                    }
                }

                return was_stanza_found;
            },

            does_stanza_property_exist: function does_stanza_property_exist(
                configuration_stanza_accessor,
                property_name,
            ) {
                var was_property_found = false;

                for (const [key, value] of Object.entries(
                    configuration_stanza_accessor.properties(),
                )) {
                    if (key === property_name) {
                        was_property_found = true;
                    }
                }

                return was_property_found;
            },

            does_storage_password_exist: function does_storage_password_exist(
                storage_passwords_accessor,
                realm,
                username,
            ) {
                var was_storage_password_found = false;
                storage_passwords = storage_passwords_accessor.list();

                for (var index = 0; index < storage_passwords.length; index++) {
                    if (storage_passwords[index].name === realm + ":" + username + ":") {
                        was_storage_password_found = true;
                    }
                }

                return was_storage_password_found;
            },

            // ---------------------
            // Retrieval Functions
            // ---------------------
            get_configuration_file: function get_configuration_file(
                configurations_accessor,
                configuration_file_name,
            ) {
                var configuration_file_accessor = configurations_accessor.item(
                    configuration_file_name,
                    {
                        // Name space information not provided
                    },
                );

                return configuration_file_accessor;
            },

            get_configuration_file_stanza: function get_configuration_file_stanza(
                configuration_file_accessor,
                configuration_stanza_name,
            ) {
                var configuration_stanza_accessor = configuration_file_accessor.item(
                    configuration_stanza_name,
                    {
                        // Name space information not provided
                    },
                );

                return configuration_stanza_accessor;
            },

            get_configuration_file_stanza_property: function get_configuration_file_stanza_property(
                configuration_file_accessor,
                configuration_file_name,
            ) {
                return null;
            },

            // ---------------------
            // Creation Functions
            // ---------------------
            create_splunk_js_sdk_service: function create_splunk_js_sdk_service(
                splunk_js_sdk,
                application_name_space,
            ) {
                var http = new splunk_js_sdk.SplunkWebHttp();

                var splunk_js_sdk_service = new splunk_js_sdk.Service(
                    http,
                    application_name_space,
                );

                return splunk_js_sdk_service;
            },

            create_configuration_file: function create_configuration_file(
                configurations_accessor,
                configuration_file_name,
            ) {
                var parent_context = this;

                return configurations_accessor.create(configuration_file_name, function(
                    error_response,
                    created_file,
                ) {
                    // Do nothing
                });
            },

            create_stanza: function create_stanza(
                configuration_file_accessor,
                new_stanza_name,
            ) {
                var parent_context = this;

                return configuration_file_accessor.create(new_stanza_name, function(
                    error_response,
                    created_stanza,
                ) {
                    // Do nothing
                });
            },

            update_stanza_properties: function update_stanza_properties(
                configuration_stanza_accessor,
                new_stanza_properties,
            ) {
                var parent_context = this;

                return configuration_stanza_accessor.update(
                    new_stanza_properties,
                    function(error_response, entity) {
                        // Do nothing
                    },
                );
            },

            create_storage_password_stanza: function create_storage_password_stanza(
                splunk_js_sdk_service_storage_passwords,
                realm,
                username,
                value_to_encrypt,
            ) {
                var parent_context = this;

                return splunk_js_sdk_service_storage_passwords.create(
                    {
                        name: username,
                        password: value_to_encrypt,
                        realm: realm,
                    },
                    function(error_response, response) {
                        // Do nothing
                    },
                );
            },

            // ----------------------------------
            // Deletion Methods
            // ----------------------------------
            delete_storage_password: function delete_storage_password(
                storage_passwords_accessor,
                realm,
                username,
            ) {
                return storage_passwords_accessor.del(realm + ":" + username + ":");
            },

            // ----------------------------------
            // Input Cleaning and Checking
            // ----------------------------------
            sanitize_string: function sanitize_string(string_to_sanitize) {
                var sanitized_string = string_to_sanitize.trim();

                return sanitized_string;
            },

            validate_api_key: function validate_api_key(api_key) {
                var error_messages = [];

                match = api_key.match('^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$');
                if (match === null) {
                  error_messages.push("Please provide a valid API Key.");
                }

                return error_messages;
            },

            validate_region: function validate_region(region) {
                var error_messages = [];

                if (region != "eu" && region != "us") {
                    error_messages.push("Region should be either us or eu");
                }

                return error_messages;
            },

            validate_inputs: function validate_inputs(api_key, region) {
                var error_messages = [];

                var api_key_errors = this.validate_api_key(api_key);
                var region_errors = this.validate_region(region);

                error_messages = error_messages.concat(api_key_errors);
                error_messages = error_messages.concat(region_errors);

                return error_messages;
            },

            // ----------------------------------
            // GUI Helpers
            // ----------------------------------
            extract_error_messages: function extract_error_messages(error_messages) {
                // A helper function to extract error messages

                // Expects an array of messages
                // [
                //     {
                //         type: the_specific_error_type_found,
                //         text: the_specific_reason_for_the_error,
                //     },
                //     ...
                // ]

                var error_messages_to_display = [];
                for (var index = 0; index < error_messages.length; index++) {
                    error_message = error_messages[index];
                    error_message_to_display =
                        error_message.type + ": " + error_message.text;
                    error_messages_to_display.push(error_message_to_display);
                }

                return error_messages_to_display;
            },

            // if needed
            redirect_to_opsgenie_setup_page: function redirect_to_opsgenie_setup_page() {
                window.location.href = "/app/opsgenie/setup_view_dashboard?success=true";
            },

            // ----------------------------------
            // Display Functions
            // ----------------------------------
            remove_outputs: function remove_outputs() {

                var error_output_element = jquery(".setup.container .error.output");

                error_output_element.stop();
                error_output_element.fadeOut({
                    complete: function() {
                        error_output_element.html("");
                    },
                });

                var success_output_element = jquery(".setup.container .success.output");

                success_output_element.stop();
                success_output_element.fadeOut({
                    complete: function() {
                        success_output_element.html("");
                    },
                });
            },

            display_error_output: function display_error_output(error_messages) {

                var error_output_element = jquery(".setup.container .error.output");

                if (error_messages.length > 0) {
                    var new_error_output_string = "";
                    new_error_output_string += "<ul>";
                    for (var index = 0; index < error_messages.length; index++) {
                        new_error_output_string +=
                            "<li>" + error_messages[index] + "</li>";
                    }
                    new_error_output_string += "</ul>";

                    error_output_element.html(new_error_output_string);
                    error_output_element.stop();
                    error_output_element.fadeIn();
                }
            },

            display_success_output: function display_success_output() {

                var success_output_element = jquery(".setup.container .success.output");

                success_output_element.html("<ul><li>Success!</li></ul>");
                success_output_element.stop();
                success_output_element.fadeIn();
            },

            get_template: function get_template() {
                template_string =
                    "<div class='setup container'>" +
                    "    <div class='left'>" +
                    "        <div class='field api_key'>" +
                    "            <div class='title'>" +
                    "                <h3>API Key:</h3>" +
                    "                Please specify the API Key to be used." +
                    "            </div>" +
                    "            </br>" +
                    "            <div class='user_input'>" +
                    "                <div class='text'>" +
                    "                    <input type='text' name='api_key' id='api_key' placeholder='Enter API Key'></input>" +
                    "                </div>" +
                    "            </div>" +
                    "        </div>" +
                    "        <div class='realm'>" +
                    "           <div class='title'>" +
                    "               <h3>Region:</h3>" +
                    "               Please specify your region." +
                    "           </div>" +
                    "           </br>" +
                    "           <div class='btn-group' data-toggle='buttons' name='region'>" +
                    "                <label class='btn btn-primary active'>" +
                    "                    <input type='radio' name='region' id='region_us' value='us'> US" +
                    "                </label>" +
                    "                <label class='btn btn-primary'>" +
                    "                    <input type='radio' name='region' id='region_eu' value='eu'> EU" +
                    "                </label>" +
                    "           </div>" +
                    "        </div>" +
                    "        <br/>" +
                    "        <br/>" +
                    "        <div>" +
                    "            <button name='setup_button' class='setup_button'>" +
                    "                Complete Setup" +
                    "            </button>" +
                    "        </div>" +
                    "        <br/>" +
                    "        <div>" +
                    "           <div class='error output'>" +
                    "           </div>" +
                    "           <div class='success output'>" +
                    "           </div>" +
                    "        </div>" +
                    "    </div>" +
                    "</div>";

                return template_string;
            },
        }); // End of ExampleView class declaration

        return ExampleView;
    }, // End of require asynchronous module definition function
); // End of require statement