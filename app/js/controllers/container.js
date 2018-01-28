angular.module('app').controller('container', ['$scope', '$routeParams', function ($scope, $routeParams) {
	$scope.callback = false;

    if ($routeParams.callback) {
        $scope.callback = $routeParams.callback;
    }
	
	$scope.container_id = $routeParams.container_id;
}]);
