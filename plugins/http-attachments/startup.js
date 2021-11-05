/*\
  title: $:/plugins/satran/http-attachments/startup.js
  type: application/javascript
  module-type: startup

  Startup initialisation

  \*/
(function(){
    /*jslint node: true, browser: true */
    /*global $tw: false */
    'use strict';
    var ENABLE_EXTERNAL_ATTACHMENTS_TITLE = '$:/config/HTTPAttachments/Enable';
    // Export name and synchronous status
    exports.name = 'http-attachments';
    exports.platforms = [
        'browser'
    ];
    exports.after = [
        'startup'
    ];
    exports.synchronous = true;
    exports.startup = function () {
        $tw.hooks.addHook('th-importing-file', function (info) {
            if ((document.location.protocol === 'http:' || document.location.protocol === 'https:') && info.isBinary && info.file.name && $tw.wiki.getTiddlerText(ENABLE_EXTERNAL_ATTACHMENTS_TITLE, '') === 'yes') {
                console.log('Wiki location', document.location.pathname)
                console.log('File location', info.file.name)
                console.log(info);
                let formData = new FormData();
                formData.append('name', info.file.name);
                formData.append('file', info.file);
                fetch('/' + info.file.name, {
                    method: 'POST',
                    body: formData
                }).then(result => {
		    info.callback([
			{
			    title: info.file.name,
			    type: info.type,
			    '_canonical_uri': info.file.name
			}
                    ]);
		}).catch(error => {
		    console.error('Error:', error);
		});
                    
                return true;
            } else {
                return false;
            }
        });
    };
})();
