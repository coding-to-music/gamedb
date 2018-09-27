if ($('#stats-page').length > 0) {

    Highcharts.chart('scores', {
        chart: {
            type: 'heatmap'
        },
        title: {
            text: ''
        },
        subtitle: {
            text: ''
        },
        xAxis: {
            categories: [''],
            title: {
                text: ''
            }
        },
        yAxis: {
            categories: [''],
            title: {
                text: ''
            }
        },
        credits: {
            enabled: false
        },
        colorAxis: {
            min: 0,
            minColor: '#FFFFFF',
            maxColor: '#28a745'
        },
        legend: {
            enabled: false
        },
        tooltip: {
            formatter: function () {
                return this.point.value + ' apps have ' + this.point.x + '/100';
            }
        },
        series: [{
            name: '',
            borderWidth: 0,
            color: '#000',
            data: scores,
            dataLabels: {
                enabled: false,
                color: '#000000'
            }
        }]
    });
}
