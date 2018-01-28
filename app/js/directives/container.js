angular.module('uber', []).directive('uberContainer', [function () {
    return {
        restrict : 'E',
        scope : {
            container_id : '=containerid',
            size : '=',
            label : '=',
            callback : '='
        },
        templateUrl : 'views/container.html',
        controller : ['$scope', '$element', '$http', '$q', function ($scope, $element, $http, $q) {

            var callback_url = $scope.callback;

            function getContainerInfo () {
                var defer = $q.defer();

                $http({
                    method : 'GET',
                    url : '/api/containers/'+$scope.container_id
                }).success(function (data, status, headers, config) {
                    defer.resolve(data);
                }).error(function (data, status, headers, config) {
                    defer.reject(status);
                });

                return defer.promise;
            }

            function uploadFile (file) {
                var formData = new FormData();

                formData.append("file", file);

                if (callback_url) {
                    formData.append('callback', callback_url);
                }

                var xhr = new XMLHttpRequest();

                xhr.addEventListener('progress', uploadProgress, false);
                xhr.addEventListener('load', uploadComplete, false);
                xhr.addEventListener('error', uploadFailed, false);

                var postUrl;
                if ($scope.newContainer) {
                    postUrl = '/api/containers';
                } else {
                    postUrl = '/api/containers/'+$scope.container_id;
                }

                xhr.open('POST', postUrl, true);
                xhr.setRequestHeader('Filename', file.name);

                xhr.onload = function () {
                    var body = JSON.parse(xhr.responseText);
                    $scope.container_id = body.container_id;
                    if ($scope.newContainer) {
                        window.location = '/#/containers/'+$scope.container_id;
                    } else {
                        loadPreview($scope.container_id);
                    }
                };

                xhr.send(formData);
            }

            function triggerFileDialog () {
                var fileInput = document.getElementById('fileInput');
                fileInput.click();
            }

            $scope.largePreview = function () {
                window.open('/#/containers/'+$scope.container_id);
            };

            $scope.downloadFile = function () {
                window.location = '/api/containers/'+$scope.container_id+'/file';
            };

            function uploadProgress (e) {
                console.log(e);
            }

            function uploadComplete (e) {
                console.log(e);
            }

            function uploadFailed (e) {
                console.log(e);
            }

            $scope.confirmDelete = function () {
                var confirmed = confirm("Are you sure you want to delete this file?");
                if (confirmed) {
                    deleteFile();
                }
            };

            function deleteFile () {

                $http({
                    method : 'DELETE',
                    url : '/api/containers/'+$scope.container_id+'/file'
                }).success(function (data, status, headers, config) {
                    console.log(data);
                    $scope.containerPreview = "";
                    $scope.emptyContainer = true;
                }).error(function (data, status, headers, config) {

                });
            }

            function loadPreview (id) {
                var img = new Image();

                if (typeof $scope.size === 'undefined') {
                    $scope.size = 900;
                }

                var imgUrl = '/api/containers/'+$scope.container_id+'/preview/'+$scope.size+'?t='+new Date().getTime();

                img.src = imgUrl;

                img.onload = function () {
                    $scope.$apply(function () {
                        $scope.containerPreview = imgUrl;
                        $scope.loading = false;
                    });
                };

                img.onerror = function () {
                    $scope.$apply(function () {
                        $scope.loading = false;
                        $scope.errorLoading = true;
                    });
                };
            }

            $scope.triggerFileDialog = function () {

                var fileInput = $element[0].querySelector('.fileInput');

                fileInput.click();

                angular.element(fileInput).on('change', function (e) {
                    var file = fileInput.files[0];

                    uploadFile(file);
                });
            };

            if (!$scope.container_id) {
                $scope.loading = false;
                $scope.newContainer = true;
            } else {
                getContainerInfo().then(function (data) {
                    console.log(data);
                    $scope.container = data;

                    if (!$scope.container.empty) {
                        loadPreview($scope.container_id);
                    }

                    $scope.loading = false;
                }, function (errCode) {
                    if (errCode === 404) {

                        $scope.newContainer = true;
                    }

                    $scope.loading = false;
                });
            }

            $element.on('dragover', function (e) {
                e.preventDefault();
                $scope.$apply(function () {
                    $scope.dropTarget = true;
                });
            });

            $element.on('dragenter', function (e) {
                e.preventDefault();
            });

            $element.on('drop', function (e) {
                e.preventDefault();

                $scope.$apply(function () {
                    $scope.dropTarget = false;
                });

                var files = e.dataTransfer.files;

                var file = files[0];

                if ($scope.newContainer) {
                    uploadFile(file);
                } else {
                    var confirmed = confirm("Do you want to replace the existing file?");
                    if (confirmed) {
                        uploadFile(file);
                    }
                }
            });

        }],
        link : function ($scope, element, attr) {
            $scope.container = false;
            $scope.loading = ($scope.container_id);
            $scope.newContainer = (!$scope.container_id);
            $scope.containerPreview = "";
            $scope.errorLoading = false;
            $scope.showInfo = false;
            $scope.dropTarget = false;

            element.on('mouseenter', function () {
                $scope.$apply(function () {
                    $scope.showInfo = true;
                });
            });

            element.on('mouseleave', function () {
                $scope.$apply(function () {
                    $scope.showInfo = false;
                    $scope.dropTarget = false;
                });
            });
        }
    };
}]);
