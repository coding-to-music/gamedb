if ($('#app-page').length > 0) {

    // Background
    var background = $('.container[data-bg]').attr('data-bg');
    if (background !== '') {
        $('body').css("background-image", 'url(' + background + ')').css("background-color", '#1b2738');
        $('nav.navbar').removeClass('bg-dark');
    }

    // News
    var $collapseBoxes = $('#news .collapse');

    $collapseBoxes.collapse();
    $collapseBoxes.first().collapse('show');

    // Price change chart
    Highcharts.chart('chart', {
        chart: {
            zoomType: 'x'
        },
        title: {text: ''},
        subtitle: {text: ''},
        xAxis: {
            title: {
                text: 'Date'
            },
            type: 'datetime'
        },
        yAxis: {
            title: {
                text: 'Price'
            },
            type: 'linear',
            min: 0,
            allowDecimals: true
        },
        legend: {
            enabled: false
        },
        credits: {
            enabled: false
        },
        series: [
            {
                type: 'line',
                name: 'Price',
                data: prices,
                step: 'right',
                color: '#28a745'
            }],
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
        }]
    });
}
