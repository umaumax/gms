(function() {
	// [Javascriptで文字列の０埋め、空白で右寄せでフォーマット - それマグで！]( http://takuya-1st.hatenablog.jp/entry/2014/12/03/114154 )
	Number.prototype.format = function(char, cnt){
		return (Array(cnt).fill(char).join("") + this.valueOf()).substr(-1*cnt); 
	};

	var fileList = [];
	//	NOTE: load data from server
	var pathname = window.location.pathname;
	console.log("pathname", pathname);
	var jqxhr = $.get( "/_api"+pathname, function(data) {
		console.log("recv data:", data);
		var names = data["names"];
		for (var i = 0; i < names.length; i++) {
			var name = names[i];
			fileList.push({text:name,seen:true});
		}
	})
	.done(function() {
		console.log("done");
	})
	.fail(function() {
		console.log("fail");
	})
	.always(function() {
		console.log("always");
	});
	Vue.component('file-item', {
		props: ['file'],
		template: '<li v-if="file.seen"><a v-bind:href="file.text" v-if="file.seen">{{ file.text }}</a></li>'
	});

	var initVue = function() {
		var fileLinkApp = new Vue({
			el: '#fileListId',
			data: {
				fileList:fileList,
			},
			methods: {
				focus : function() {
					$('#grepInput input').focus();
				}
			}
		});

		var grepInputApp = new Vue({
			el: '#grepInput',
			data: {
				message: 'How to use!!\n\n1. input search keyword!\n2. tab (+shift)\n3. enter with (command[tab] or shift[window])\n\n A. in file list key "esc" or delete -> refocus',
			caseInsensitiveChecked:true,
				regexpChecked:true,
				keyword: '',
				errorMessage: ''
			},
			methods: {
				grep: function () {
					var fileList = fileLinkApp.fileList;
					//	NOTE to avoid display:none
					this.message=' ';
					this.errorMessage = '';

					var keyword = this.keyword.toLowerCase();
					//				console.log("keyword", keyword);
					var re = null;
					if (this.regexpChecked) {
						try {
							var flags = 'mg';
							if (this.caseInsensitiveChecked) flags += 'i';
							re = new RegExp(keyword, flags);
						}catch(e) {
							this.errorMessage=String(e);
						}
						//				console.log("RegExp", re);
						//						if (re == null) {
						//							return;
						//						}
					}

					if (re!=null && this.regexpChecked) {
						$("#fileListId").unmark();
						$("#fileListId").markRegExp(re);
					}else {
						$("#fileListId").unmark();
						$("#fileListId").mark(keyword);
					}

					// NOTE title change example
					//	$("h1").unmark();
					//	$("h1").mark(this.message);
					//	$("h1").markRegExp(re);

					var cnt = 0;
					for (var i = 0; i < fileList.length; i++) {
						var ret=false;
						var target = fileList[i].text;
						// NOTE:正規表現が無効な場合には通常の検索も行う
						if (re!=null && this.regexpChecked) {
							// NOTE:bugあり?もしくは仕様が想定と異なる?
							//							ret = re.test(target);
							ret = target.match(re)!=null;
						}
						if (!ret) {
							ret = target.toLowerCase().includes(keyword);
						}
						if (ret) {
							cnt++;
							fileList[i].seen=true;
						}else {
							fileList[i].seen=false;
						}
					}

					this.message = (cnt).format(" ",3) + ' Hit!!';
				}
			}
		});
		//	NOTE this function must be called after new Vue()
		$('#grepInput input').focus();
	};

	//	<link rel="import" href="/_assets/vue-app.html" onload="handleLoad(event)" onerror="handleError(event)">
	var link = document.createElement('link');
	link.rel = 'import';
	link.href = '/_assets/vue-app.html'
	link.onload = function handleLoad(e) {
		var link = document.querySelector('link[rel="import"]');
		var content = link.import;

		var el = content.querySelector('.extended_html');

		document.body.appendChild(el.cloneNode(true));

		console.log('Loaded import: ' + e.target.href);
		initVue();
	};
	link.onerror = function handleError(e) {
		console.log('Error loading import: ' + e.target.href);
	}
	document.head.appendChild(link);
}).call(this);
