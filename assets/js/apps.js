if ($('#apps-page').length > 0) {

    $('select.form-control-chosen').chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false
    });
}

if ($('#app-page').length > 0) {

    $('.collapse').collapse();

    var x = [
        [
            1370131200000,
            0.7695
        ],
        [
            1370217600000,
            0.7648
        ],
        [
            1370304000000,
            0.7645
        ],
        [
            1370390400000,
            0.7638
        ]
    ];

    Highcharts.chart('container', {
        chart: {
            zoomType: 'x'
        },
        title: {
            text: 'USD to EUR exchange rate over time'
        },
        subtitle: {
            text: document.ontouchstart === undefined ?
                'Click and drag in the plot area to zoom in' : 'Pinch the chart to zoom in'
        },
        xAxis: {
            type: 'datetime'
        },
        yAxis: {
            title: {
                text: 'Exchange rate'
            }
        },
        legend: {
            enabled: false
        },
        annotations: [{
            labelOptions: {
                backgroundColor: 'rgba(255,255,255,0.5)',
                verticalAlign: 'top',
                y: 15
            },
            labels: [{
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 27.98,
                    y: 255
                },
                text: 'Arbois'
            }, {
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 45.5,
                    y: 611
                },
                text: 'Montrond'
            }, {
                point: {
                    xAxis: 0,
                    yAxis: 0,
                    x: 63,
                    y: 651
                },
                text: 'Mont-sur-Monnet'
            }]
        }],

        plotOptions: {
            area: {
                fillColor: {
                    linearGradient: {
                        x1: 0,
                        y1: 0,
                        x2: 0,
                        y2: 1
                    },
                    stops: [
                        [0, Highcharts.getOptions().colors[0]],
                        [1, Highcharts.Color(Highcharts.getOptions().colors[0]).setOpacity(0).get('rgba')]
                    ]
                },
                marker: {
                    radius: 2
                },
                lineWidth: 1,
                states: {
                    hover: {
                        lineWidth: 1
                    }
                },
                threshold: null
            }
        },

        series: [{
            type: 'area',
            name: 'USD to EUR',
            data: x,
            step: true
        }]
    });
}
