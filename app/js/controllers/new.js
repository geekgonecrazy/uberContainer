angular.module('app').controller('new_container', ['$scope', '$routeParams', function ($scope, $routeParams) {
    $scope.callback = false;
    
    if ($routeParams.callback) {
        $scope.callback = $routeParams.callback;
    }

	$scope.container_id = false;
}]);
