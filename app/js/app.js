var app = angular.module('app', ['ngRoute', 'uber']);

app.config(['$controllerProvider', '$routeProvider', '$locationProvider', function ($controllerProvider, $routeProvider, $locationProvider) {
	app.controller = $controllerProvider.register;

	$routeProvider.when('/new', {
		templateUrl : '/views/container_view.html',
		controller : 'new_container'
	}).when('/containers/:container_id', {
		templateUrl : '/views/container_view.html',
		controller : 'container'
	}).otherwise({
		redirectTo: '/new'
	});
}]);
